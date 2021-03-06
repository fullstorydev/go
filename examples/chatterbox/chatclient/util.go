package chatclient

import (
	"context"
	"fmt"
	"io"

	"github.com/fullstorydev/go/examples/chatterbox"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// filterClientError cleans up error logging by filtering out errors related to (probably user initiated) cancel.
func filterClientError(err error) error {
	if err == io.EOF || err == context.Canceled {
		return nil
	}
	if code := status.Code(err); code == codes.Canceled {
		return nil
	}
	return err
}

// commonClientStream intersects ChatterBox_ChatClient and ChatterBox_MonitorClient
type commonClientStream interface {
	grpc.ClientStream
	Recv() (*chatterbox.Event, error)
}

// fetchInitialState ensures we read a complete initial model from the server
func fetchInitialState(ctx context.Context, stream commonClientStream) (chatterbox.MembersModel, error) {
	// Wait for the initial state to come back.
	members := chatterbox.MembersModel{}
	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("stream.Recv: %w", err)
		}
		switch msg.What {
		case chatterbox.What_INITIALIZED:
			return members, nil
		case chatterbox.What_JOIN:
			members.Add(msg.Who)
		default:
			return nil, fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}
