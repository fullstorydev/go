package chatterbox

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// filterErr cleans up error logging by filtering out errors related to (probably user initiated) cancel.
func filterErr(err error) error {
	if err == io.EOF || err == context.Canceled {
		return nil
	}
	if code := status.Code(err); code == codes.Canceled {
		return nil
	}
	return err
}

// commonStream is an intesection type for ChatterBox_ChatClient and ChatterBox_MonitorClient
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
