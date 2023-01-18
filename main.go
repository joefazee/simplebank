package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/joefazee/simplebank/doc/swagger"
	"github.com/joefazee/simplebank/gapi"
	"github.com/joefazee/simplebank/pb"
	"github.com/joefazee/simplebank/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"

	"github.com/joefazee/simplebank/api"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/util"
	_ "github.com/lib/pq"
)

const (
	development = "development"
)

func main() {

	config, err := util.LoadCondfig(".")

	if err != nil {
		log.Fatal().Err(err)
	}

	if config.Environment == development {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err)
	}

	runDBMigration(config.DbMigrationURL, config.DBSource)

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{Addr: config.RedisAddress}
	taskDistributor := worker.NewRedisTaskTaskDistributor(redisOpt)

	go runTaskProcessor(redisOpt, store)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)

}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal().Msgf("migration error %s", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("migration error %s", err)
	}

	log.Info().Msg("migration successful")
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msg("unable to create grpc server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatal().Msgf("unable to listen on %s (%s)", config.GrpcServerAddress, err.Error())
	}

	log.Printf("starting grpc server on %s", config.GrpcServerAddress)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msgf("unable to serve request on %s (%s)", config.GrpcServerAddress, err.Error())
	}

}

func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Msgf("unable to create grpc server %s", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msgf("unable to register server handler %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.FS(swagger.Doc))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HttServerAddress)
	if err != nil {
		log.Fatal().Msgf("unable to listen on %s (%s)", config.HttServerAddress, err.Error())
	}

	log.Printf("starting HTTP Gateway server on %s", config.HttServerAddress)
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msgf("unable to serve request on %s (%s)", config.HttServerAddress, err.Error())
	}

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msgf("unable to create server")
	}

	err = server.Start(config.HttServerAddress)

	if err != nil {
		log.Fatal().Err(err)
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) {
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store)
	log.Info().Msg("starting task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start task processor")
	}
}
