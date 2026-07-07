package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ats-tech/cineforge/services/media-worker/internal/processor"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	ctx := context.Background()
	p, err := processor.New(ctx, required("DATABASE_URL"), os.Getenv("GCS_ENDPOINT"), required("GCS_BUCKET_QUARANTINE"), required("GCS_BUCKET_MEDIA"))
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	tc, err := client.Dial(client.Options{HostPort: required("TEMPORAL_ADDRESS")})
	if err != nil {
		log.Fatal(err)
	}
	defer tc.Close()
	w := worker.New(tc, "cineforge-media", worker.Options{})
	w.RegisterWorkflowWithOptions(IngestMedia, workflow.RegisterOptions{Name: "IngestMedia"})
	w.RegisterActivity(processor.Activities{Processor: p})
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
func IngestMedia(ctx workflow.Context, assetID string) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: 2 * time.Hour, HeartbeatTimeout: 2 * time.Minute, RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 3, InitialInterval: 10 * time.Second}}
	ctx = workflow.WithActivityOptions(ctx, opts)
	return workflow.ExecuteActivity(ctx, "ProcessMedia", assetID).Get(ctx, nil)
}
func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}
