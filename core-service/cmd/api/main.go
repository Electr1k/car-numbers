package main

import (
	"context"
	"core-service/internal/app"
	httptransport "core-service/internal/http"
	platehandler "core-service/internal/http/handler/plate"
	regionhandler "core-service/internal/http/handler/region"
	valuationhandler "core-service/internal/http/handler/valuation"
	"core-service/internal/service/ml"
	"core-service/internal/service/plate"
	"core-service/internal/usecase/fetchplate"
	"core-service/internal/usecase/fetchplates"
	"core-service/internal/usecase/fetchregions"
	"core-service/internal/usecase/valuation"
)

func main() {
	app.Run("api", func(ctx context.Context, a *app.App) error {
		plateClient := plate.NewClient(a.Config.PlateConfig, a.Logger)
		mlClient := ml.NewClient(a.Config.MLConfig, a.Logger)
		router := httptransport.NewRouter(a.Config.HTTPServer, a.Logger, httptransport.Handlers{
			Plate:     platehandler.New(fetchplates.New(plateClient), fetchplate.New(plateClient)),
			Regions:   regionhandler.New(fetchregions.New(plateClient)),
			Valuation: valuationhandler.New(valuation.New(mlClient, plateClient, a.Logger)),
		})
		httpServer := httptransport.NewServer(a.Config.HTTPServer, router)

		return httpServer.Run(ctx)
	})
}
