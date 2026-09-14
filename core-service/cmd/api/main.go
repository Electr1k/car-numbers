package main

import (
	"context"
	"core-service/internal/app"
	httptransport "core-service/internal/http"
	platehandler "core-service/internal/http/handler/plate"
	"core-service/internal/service/plate"
	"core-service/internal/usecase/fetchplates"
)

func main() {
	app.Run("api", func(ctx context.Context, a *app.App) error {
		plateClient := plate.NewClient(a.Config.PlateConfig, a.Logger)
		router := httptransport.NewRouter(a.Config.HTTPServer, a.Logger, httptransport.Handlers{
			Plate: platehandler.New(fetchplates.New(plateClient)),
		})
		httpServer := httptransport.NewServer(a.Config.HTTPServer, router)

		return httpServer.Run(ctx)
	})
}
