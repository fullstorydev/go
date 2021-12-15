package chatterbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

// RunListener is an example of a gRPC client for a one-way server stream.
func RunListener(ctx context.Context, cl ChatterBoxClient) error {
	const maxBackoff = 16 * time.Second
	backoff := time.Second

	// Loop forever until killed or cancelled.
	for {
		ok, err := listenIterate(ctx, cl)
		if err != nil {
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

func listenIterate(ctx context.Context, cl ChatterBoxClient) (bool, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := cl.Listen(ctx, &emptypb.Empty{})
	if err != nil {
		return false, fmt.Errorf("cl.Listen: %w", err)
	}

	members, err := fetchInitialState(ctx, stream)
	if err != nil {
		return false, fmt.Errorf("fetchInitialState: %w", err)
	}

	// Successfully fetched initial state; from here on return true to reset backoff.
	log.Printf("Members: %+v", members)

	// Monitor the connection.
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
