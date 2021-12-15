package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fullstorydev/go/examples/chatterbox"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	switch flag.Arg(0) {
	case "":
		err = fmt.Errorf(`choose one of: "server", "client", "listen" `)
	case "server":
		err = runServer(ctx)
	case "client":
		err = runClient(ctx)
	case "listen":
		err = runListener(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", flag.Arg(0))
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runServer(ctx context.Context) error {
	svr := grpc.NewServer()
	chatterbox.RegisterChatterBoxServer(svr, chatterbox.NewServer())

	lis, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Println("Listening on port 9000")
	return svr.Serve(lis)
}

func runClient(ctx context.Context) error {
	log.Println("Dialing port 9000")
	conn, err := grpc.DialContext(ctx, "127.0.0.1:9000", grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	return chatterbox.RunClient(ctx, chatterbox.NewChatterBoxClient(conn))
}

func runListener(ctx context.Context) error {
	log.Println("Dialing port 9000")
	conn, err := grpc.DialContext(ctx, "127.0.0.1:9000", grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	return chatterbox.RunListener(ctx, chatterbox.NewChatterBoxClient(conn))
}
