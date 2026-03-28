package main

import (
	"context"
	"os/signal"
	"syscall"

	"IFMS-BE-API/internal/app/api"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := api.NewApiApplication(ctx)
	application.Start()
	defer application.Shutdown()

	<-ctx.Done()
}
