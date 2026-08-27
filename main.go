package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/billiraheem/Billi-Bank/api"
	db "github.com/billiraheem/Billi-Bank/db/sqlc"
	"github.com/billiraheem/Billi-Bank/gapi"
	"github.com/billiraheem/Billi-Bank/pb"
	"github.com/billiraheem/Billi-Bank/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	// run db migrations directly in the code this is for our docker image
	runDBMigrations(config.MigrationURL, config.DBSource)

	runGinServer(config, store)
	// go runGatewayServer(config, store) // runs in a seperate gorountine so the 2 servers don't block each other
	// runGrpcServer(config, store)
}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil{
		log.Fatal("cannot start server:", err)
	}
}

func runGrpcServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBilliBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil{
		log.Fatal("cannot start listener:", err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil{
		log.Fatal("cannot start gRPC server:", err)
	}
}

// serves both http and grpc requests at the same time
func runGatewayServer(config utils.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
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

	err = pb.RegisterBilliBankHandlerServer(ctx, grpcMux, server)
	if err != nil{
		log.Fatal("cannot register handler server:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.ServerAddress)
	if err != nil{
		log.Fatal("cannot start listener:", err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil{
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

func runDBMigrations(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create a migration instance:", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up: ", err)
	}

	log.Println("db migration successful")
}