package processor

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/image/webp"
	"google.golang.org/api/option"
)

type Processor struct {
	db                            *pgxpool.Pool
	storage                       *storage.Client
	quarantineBucket, mediaBucket string
}
type asset struct {
	ID, WorkspaceID, Kind, Filename, ContentType, Object string
	Size                                                 int64
}

func New(ctx context.Context, databaseURL, endpoint, quarantineBucket, mediaBucket string) (*Processor, error) {
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	opts := []option.ClientOption{}
	if endpoint != "" {
		opts = append(opts, option.WithEndpoint(endpoint), option.WithoutAuthentication())
	}
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Processor{db: db, storage: client, quarantineBucket: quarantineBucket, mediaBucket: mediaBucket}, nil
}
func (p *Processor) Close() { p.db.Close(); _ = p.storage.Close() }

func (p *Processor) Process(ctx context.Context, assetID string) error {
	var a asset
	if err := p.db.QueryRow(ctx, `update media_assets set state='scanning',updated_at=now() where id=$1 and state='uploaded' returning id::text,workspace_id::text,kind,original_filename,content_type,size_bytes,quarantine_object`, assetID).Scan(&a.ID, &a.WorkspaceID, &a.Kind, &a.Filename, &a.ContentType, &a.Size, &a.Object); err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "cineforge-media-*")
	if err != nil {
		return p.reject(ctx, a.ID, "temporary_storage_failed")
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "input"+filepath.Ext(a.Filename))
	if err := p.download(ctx, a.Object, input); err != nil {
		return p.reject(ctx, a.ID, "quarantine_read_failed")
	}
	if err := scanMalware(ctx, input); err != nil {
		return p.reject(ctx, a.ID, "malware_detected")
	}
	metadata, err := inspect(ctx, a.Kind, input)
	if err != nil {
		return p.reject(ctx, a.ID, "invalid_media")
	}
	if _, err := p.db.Exec(ctx, `update media_assets set state='processing',updated_at=now() where id=$1`, a.ID); err != nil {
		return err
	}
	checksum, err := checksumFile(input)
	if err != nil {
		return p.reject(ctx, a.ID, "checksum_failed")
	}
	original := fmt.Sprintf("originals/%s/%s/%s", a.WorkspaceID, a.ID, safeName(a.Filename))
	if err := p.upload(ctx, original, input, a.ContentType); err != nil {
		return p.reject(ctx, a.ID, "storage_write_failed")
	}
	preview, thumb, waveform := p.derivatives(ctx, a, dir, input)
	b, _ := json.Marshal(metadata)
	vid := uuid.Must(uuid.NewV7()).String()
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `insert into media_asset_versions(id,workspace_id,asset_id,version,storage_object,preview_object,thumbnail_object,waveform_object,checksum_sha256,metadata) values($1,$2,$3,1,$4,$5,$6,$7,$8,$9)`, vid, a.WorkspaceID, a.ID, original, nullString(preview), nullString(thumb), nullString(waveform), checksum, b); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `update media_assets set state='ready',updated_at=now() where id=$1`, a.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Processor) download(ctx context.Context, object, path string) error {
	r, err := p.storage.Bucket(p.quarantineBucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(r, 2<<30+1))
	return err
}
func (p *Processor) upload(ctx context.Context, object, path, contentType string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := p.storage.Bucket(p.mediaBucket).Object(object).NewWriter(ctx)
	w.ContentType = contentType
	w.CacheControl = "private, max-age=31536000, immutable"
	if _, err = io.Copy(w, f); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
func (p *Processor) reject(ctx context.Context, id, code string) error {
	_, err := p.db.Exec(ctx, `update media_assets set state='rejected',updated_at=now() where id=$1`, id)
	if err != nil {
		return err
	}
	return fmt.Errorf("media rejected: %s", code)
}

func scanMalware(ctx context.Context, path string) error {
	addr := os.Getenv("CLAMD_HOST")
	if addr == "" {
		addr = "localhost:3310"
	}
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Minute))
	if _, err = conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, 1<<20)
	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			header := make([]byte, 4)
			binary.BigEndian.PutUint32(header, uint32(n))
			if _, err = conn.Write(header); err != nil {
				return err
			}
			if _, err = conn.Write(buf[:n]); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if _, err = conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return err
	}
	response, err := io.ReadAll(io.LimitReader(conn, 4096))
	if err != nil {
		return err
	}
	if !strings.Contains(string(response), "OK") {
		return fmt.Errorf("clamav rejected file")
	}
	return nil
}
func inspect(ctx context.Context, kind, path string) (map[string]any, error) {
	if kind == "image" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		cfg, format, err := image.DecodeConfig(f)
		if err != nil {
			if _, seekErr := f.Seek(0, 0); seekErr == nil {
				cfg, err = webp.DecodeConfig(f)
				format = "webp"
			}
		}
		if err != nil || int64(cfg.Width)*int64(cfg.Height) > 100_000_000 {
			return nil, fmt.Errorf("invalid image")
		}
		return map[string]any{"width": cfg.Width, "height": cfg.Height, "format": format}, nil
	}
	if kind == "3d" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		header := make([]byte, 12)
		if _, err = io.ReadFull(f, header); err != nil || string(header[:4]) != "glTF" {
			return nil, fmt.Errorf("invalid glb")
		}
		return map[string]any{"format": "glb"}, nil
	}
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_format", "-show_streams", "-of", "json", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err = json.Unmarshal(out, &data); err != nil {
		return nil, err
	}
	return data, nil
}
func (p *Processor) derivatives(ctx context.Context, a asset, dir, input string) (preview, thumb, waveform string) {
	base := fmt.Sprintf("derived/%s/%s", a.WorkspaceID, a.ID)
	if a.Kind == "video" {
		pp := filepath.Join(dir, "preview.mp4")
		if exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-y", "-i", input, "-vf", "scale='min(1280,iw)':-2", "-c:v", "libx264", "-preset", "veryfast", "-movflags", "+faststart", "-c:a", "aac", pp).Run() == nil {
			preview = base + "/preview.mp4"
			_ = p.upload(ctx, preview, pp, "video/mp4")
		}
		tp := filepath.Join(dir, "thumb.jpg")
		if exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-y", "-ss", "00:00:01", "-i", input, "-frames:v", "1", "-vf", "scale=640:-2", tp).Run() == nil {
			thumb = base + "/thumbnail.jpg"
			_ = p.upload(ctx, thumb, tp, "image/jpeg")
		}
	}
	if a.Kind == "audio" || a.Kind == "voice" {
		pp := filepath.Join(dir, "preview.m4a")
		if exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-y", "-i", input, "-c:a", "aac", "-ar", "48000", pp).Run() == nil {
			preview = base + "/preview.m4a"
			_ = p.upload(ctx, preview, pp, "audio/mp4")
		}
	}
	return
}
func checksumFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func safeName(v string) string {
	v = filepath.Base(v)
	v = strings.Map(func(r rune) rune {
		if r > 127 || strings.ContainsRune("\\/:*?\"<>|", r) {
			return '_'
		}
		return r
	}, v)
	if v == "" {
		return "asset"
	}
	return v
}
func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

type Activities struct{ Processor *Processor }

func (a Activities) ProcessMedia(ctx context.Context, assetID string) error {
	return a.Processor.Process(ctx, assetID)
}

var _ = time.Now
