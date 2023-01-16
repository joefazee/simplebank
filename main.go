package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joefazee/simplebank/doc/swagger"
	"github.com/joefazee/simplebank/gapi"
	"github.com/joefazee/simplebank/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"

	"github.com/joefazee/simplebank/api"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadCondfig(".")

	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	runDBMigration(config.DbMigrationURL, config.DBSource)

	store := db.NewStore(conn)

	go runGatewayServer(config, store)
	runGrpcServer(config, store)

}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Fatal("migration error", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("migration error", err)
	}

	log.Println("migration successful")
}

func runGrpcServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("unable to create grpc server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		log.Fatalf("unable to listen on %s (%s)", config.GrpcServerAddress, err.Error())
	}

	log.Printf("starting grpc server on %s", config.GrpcServerAddress)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("unable to serve request on %s (%s)", config.GrpcServerAddress, err.Error())
	}

}

func runGatewayServer(config util.Config, store db.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("unable to create grpc server", err)
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
		log.Fatal("unable to register server handler", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	fs := http.FileServer(http.FS(swagger.Doc))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	listener, err := net.Listen("tcp", config.HttServerAddress)
	if err != nil {
		log.Fatalf("unable to listen on %s (%s)", config.HttServerAddress, err.Error())
	}

	log.Printf("starting HTTP Gateway server on %s", config.HttServerAddress)
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatalf("unable to serve request on %s (%s)", config.HttServerAddress, err.Error())
	}

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("unable to create server")
	}

	err = server.Start(config.HttServerAddress)

	if err != nil {
		log.Fatal(err)
	}
}
