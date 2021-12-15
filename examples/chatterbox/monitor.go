package chatterbox

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
)

// Common logic to process the server stream (both client and listener).

// commonStream is an intesection type for ChatterBox_ChatClient and ChatterBox_ListenClient
type commonStream interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

// fetchInitialState ensures we read a complete initial model from the server
func fetchInitialState(ctx context.Context, stream commonStream) (MembersModel, error) {
	// Wait for the initial state to come back.
	members := MembersModel{}
	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("stream.Recv: %w", err)
		}
		switch msg.What {
		case What_INITIALIZED:
			return members, nil
		case What_JOIN:
			members.Add(msg.Who)
		default:
			return nil, fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}

// monitor pulls from the given stream until it closes, updating members model.
func monitor(_ context.Context, members MembersModel, stream commonStream) error {
	// Monitor the connection.
	for {
		msg, err := stream.Recv()
		if err != nil {
			return filterErr(err)
		}

		switch msg.What {
		case What_CHAT:
			log.Printf("%s: %s", msg.Who, msg.Text)
		case What_JOIN:
			members.Add(msg.Who)
			log.Printf("%s: joined", msg.Who)
			log.Printf("Members: %s", members)
		case What_LEAVE:
			members.Remove(msg.Who)
			log.Printf("%s: left", msg.Who)
			log.Printf("Members: %s", members)
		default:
			return fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}
