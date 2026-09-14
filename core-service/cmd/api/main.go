package main

import (
	"context"
	"core-service/internal/app"
	"fmt"
)

func main() {
	app.Run("api", func(ctx context.Context, a *app.App) error {

		//router := httptransport.NewRouter(a.Config.HTTPServer, a.Logger, httptransport.Handlers{
		//	Region: region.New(fetchregions.New(postgres.NewRegionRepository(a.Database))),
		//	Feed:   feed.New(fetchfeednumbers.New(offerRepository)),
		//	Number: number.New(fetchnumber.New(offerRepository)),
		//})
		//httpServer := httptransport.NewServer(a.Config.HTTPServer, router)
		fmt.Println("api start")
		return nil
		//return httpServer.Run()
	})

}
