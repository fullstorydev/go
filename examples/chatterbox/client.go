package chatterbox

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// RunClient is an example of a gRPC client for a bidi server stream.
func RunClient(ctx context.Context, cl ChatterBoxClient) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Read lines off the terminal, try to send through channel.
	msgs := make(chan string)
	go func() {
		// when exiting for any reason, cancel the stream context.
		defer cancel()
		defer close(msgs)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msgs <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}()

	const maxBackoff = 16 * time.Second
	backoff := time.Second

	// Loop forever until killed or cancelled.
	for {
		ok, err := clientIterate(ctx, msgs, cl)
		if err := filterErr(err); err != nil {
			log.Println(err)
		}
		if ok {
			backoff = time.Second
		} else {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
	}
}

func clientIterate(ctx context.Context, messages <-chan string, cl ChatterBoxClient) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := cl.Chat(ctx)
	if err != nil {
		return false, fmt.Errorf("cl.Chat: %w", err)
	}

	members, err := fetchInitialState(ctx, stream)
	if err != nil {
		return false, fmt.Errorf("fetchInitialState: %w", err)
	}

	// Successfully fetched initial state; from here on return true to reset backoff.
	log.Printf("Members: %+v", members)

	// Send chat to the server in the background.
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		// when exiting for any reason, cancel the outbound stream
		defer wg.Done()
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-messages:
				if !ok {
					return
				}
				if err := stream.Send(&Send{
					Text: msg,
				}); err != nil {
					log.Printf("Failed to send: %s", err)
					return
				}
			}
		}
	}()

	// Monitor the connection in the foreground.
	for {
		msg, err := stream.Recv()
		if err != nil {
			return true, err
		}

		switch msg.What {
		case What_CHAT:
			log.Printf("%s: %s", msg.Who, msg.Text)
		case What_JOIN:
			members[msg.Who] = struct{}{}
			log.Printf("%s: joined", msg.Who)
			log.Printf("Members: %+v", members)
		case What_LEAVE:
			delete(members, msg.Who)
			log.Printf("%s: left", msg.Who)
			log.Printf("Members: %+v", members)
		default:
			return true, fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}

type commonClient interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

// fetchInitialState ensures we read a complete initial model from the server
func fetchInitialState(ctx context.Context, stream commonClient) (map[string]struct{}, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	// Wait for the initial state to come back.
	members := map[string]struct{}{}
	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("stream.Recv: %w", err)
		}
		switch msg.What {
		case What_INITIALIZED:
			return members, nil
		case What_JOIN:
			members[msg.Who] = struct{}{}
		default:
			return nil, fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}
