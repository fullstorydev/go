package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fullstorydev/go/examples/chatterbox"
	"github.com/fullstorydev/go/examples/chatterbox/chatclient"
	"github.com/fullstorydev/go/examples/chatterbox/chatserver"
	"google.golang.org/grpc"
)

const (
	addr = "127.0.0.1:9000"
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	switch flag.Arg(0) {
	case "":
		err = fmt.Errorf(`choose one of: "server", "client", "monitor" `)
	case "server":
		err = runServer(ctx)
	case "client":
		err = runClient(ctx)
	case "monitor":
		err = runMonitor(ctx)
	default:
		err = fmt.Errorf("unknown command: %s", flag.Arg(0))
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runServer(_ context.Context) error {
	svr := grpc.NewServer()
	chatterbox.RegisterChatterBoxServer(svr, chatserver.NewServer())

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Println("Listening on ", addr)
	return svr.Serve(lis)
}

func runClient(ctx context.Context) error {
	log.Println("Dialing ", addr)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	// Read lines off the terminal, try to send through channel.
	ctx, cancel := context.WithCancel(ctx)
	chatInput := make(chan string)
	go func() {
		// when exiting for any reason, cancel the stream context.
		defer cancel()
		defer close(chatInput)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case chatInput <- scanner.Text():
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}()

	return chatclient.RunClient(ctx, chatInput, chatterbox.NewChatterBoxClient(conn))
}

func runMonitor(ctx context.Context) error {
	log.Println("Dialing ", addr)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	return chatclient.RunMonitor(ctx, chatterbox.NewChatterBoxClient(conn))
}
