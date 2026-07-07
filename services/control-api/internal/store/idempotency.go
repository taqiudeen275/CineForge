package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrIdempotencyMismatch   = errors.New("idempotency key was reused with a different request")
	ErrIdempotencyInProgress = errors.New("an operation with this idempotency key is in progress")
)

type IdempotencyReplay struct {
	Exists bool
	Status int
	Body   []byte
}

func (s *Store) BeginIdempotency(ctx context.Context, workspaceID, userID, key, route, requestHash string) (IdempotencyReplay, error) {
	tag, err := s.pool.Exec(ctx, `insert into idempotency_keys(workspace_id,actor_user_id,key,route,request_hash,expires_at) values(nullif($1,'')::uuid,$2,$3,$4,$5,$6) on conflict do nothing`, workspaceID, userID, key, route, requestHash, time.Now().Add(24*time.Hour))
	if err != nil {
		return IdempotencyReplay{}, err
	}
	if tag.RowsAffected() == 1 {
		return IdempotencyReplay{}, nil
	}
	var storedHash string
	var status *int
	var body []byte
	err = s.pool.QueryRow(ctx, `select request_hash,response_status,response_body from idempotency_keys where workspace_id is not distinct from nullif($1,'')::uuid and actor_user_id=$2 and key=$3 and route=$4 and expires_at>now()`, workspaceID, userID, key, route).Scan(&storedHash, &status, &body)
	if errors.Is(err, pgx.ErrNoRows) {
		return IdempotencyReplay{}, ErrIdempotencyInProgress
	}
	if err != nil {
		return IdempotencyReplay{}, err
	}
	if storedHash != requestHash {
		return IdempotencyReplay{}, ErrIdempotencyMismatch
	}
	if status == nil {
		return IdempotencyReplay{}, ErrIdempotencyInProgress
	}
	return IdempotencyReplay{Exists: true, Status: *status, Body: body}, nil
}

func (s *Store) CompleteIdempotency(ctx context.Context, workspaceID, userID, key, route string, status int, body []byte) error {
	_, err := s.pool.Exec(ctx, `update idempotency_keys set response_status=$1,response_body=case when $2='' then null else $2::jsonb end where workspace_id is not distinct from nullif($3,'')::uuid and actor_user_id=$4 and key=$5 and route=$6`, status, string(body), workspaceID, userID, key, route)
	return err
}

func (s *Store) AbandonIdempotency(ctx context.Context, workspaceID, userID, key, route string) {
	_, _ = s.pool.Exec(ctx, `delete from idempotency_keys where workspace_id is not distinct from nullif($1,'')::uuid and actor_user_id=$2 and key=$3 and route=$4 and response_status is null`, workspaceID, userID, key, route)
}
