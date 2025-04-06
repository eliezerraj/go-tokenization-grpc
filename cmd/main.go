package main

import(
	"time"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-tokenization-grpc/internal/infra/configuration"
	"github.com/go-tokenization-grpc/internal/core/model"
	"github.com/go-tokenization-grpc/internal/core/service"
	"github.com/go-tokenization-grpc/internal/infra/server"
	"github.com/go-tokenization-grpc/internal/adapter/database"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
	grpc_adapter "github.com/go-tokenization-grpc/internal/adapter/grpc/server"
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
	childLogger = log.With().Str("component","go-tokenization-grpc").Str("package", "main").Logger()
)

// About initialize the enviroment var
func init(){
	childLogger.Info().Str("func","init").Send()

	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv()

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.DatabaseConfig = &databaseConfig
	appServer.ConfigOTEL = &configOTEL
}

func main()  {
	childLogger.Info().Str("func","main").Interface("appServer :",appServer).Send()

	ctx := context.Background()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				childLogger.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				childLogger.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// create and wire
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database, appServer.ApiService, )
	adapterGrpc := grpc_adapter.NewAdapterGrpc(&appServer, workerService)
	workerServer := server.NewWorkerServer(adapterGrpc)

	// start grpc server
	workerServer.StartGrpcServer(ctx, &appServer)
}