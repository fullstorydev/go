package chatclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fullstorydev/go/examples/chatterbox"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Because clients run infinite retry loops, they should exponentially backoff when servers are unavailable.
	// This might be much higher in a real system
	maxBackoff = 8 * time.Second
	minBackoff = time.Second
)

// RunMonitor is an example of a gRPC client for a one-way server stream.
func RunMonitor(ctx context.Context, cl chatterbox.ChatterBoxClient) error {
	mm := &MembersMonitor{
		cl: cl,
	}

	if err := mm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	// Every 10 seconds, print the current member list.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		log.Printf("Members: %s", mm.GetMembers())

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

type MembersMonitor struct {
	cl chatterbox.ChatterBoxClient

	mu      sync.RWMutex
	members chatterbox.MembersModel
}

// GetMembers returns the current list of members (at all times) to the rest of the application.
func (mm *MembersMonitor) GetMembers() []string {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.members.Strings()
}

// Start this MembersMonitor. Fetches the initial state synchronously, then background monitors until ctx is cancelled.
func (mm *MembersMonitor) Start(ctx context.Context) error {
	started := false
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		// Ensure cleanup happens if we do not start successfully.
		if !started {
			cancel()
		}
	}()

	// Synchronously ensure we can fetch an initial model before we return.
	stream, err := mm.startStream(ctx)
	if err != nil {
		return err // failed to fetch the initial state
	}
	started = true

	go func() {
		// Monitor the first stream until it dies.
		if err := mm.monitorStream(ctx, stream); err != nil {
			log.Println(err)
		}
		if ctx.Err() != nil {
			return
		}
		// Run until ctx is cancelled.
		if err := mm.Run(ctx); err != nil {
			log.Println(err)
		}
	}()
	return nil
}

// Run runs this MembersMonitor in the foreground until ctx is cancelled.
func (mm *MembersMonitor) Run(ctx context.Context) error {
	backoff := minBackoff

	// Loop forever until killed or cancelled.
	for {
		ok, err := func() (bool, error) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			stream, err := mm.startStream(ctx)
			if err != nil {
				return false, err
			}

			return true, mm.monitorStream(ctx, stream)
		}()
		if err != nil {
			log.Println(err)
		}

		if ok {
			backoff = minBackoff
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

func (mm *MembersMonitor) startStream(ctx context.Context) (chatterbox.ChatterBox_MonitorClient, error) {
	stream, err := mm.cl.Monitor(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("cl.Monitor: %w", err)
	}

	members, err := fetchInitialState(ctx, stream)
	if err != nil {
		return nil, fmt.Errorf("fetchInitialState: %w", err)
	}

	// Successfully fetched initial state.
	func() {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		mm.members = members
	}()
	return stream, nil
}

// monitorStream pulls from the given stream until it closes, updating the members model.
func (mm *MembersMonitor) monitorStream(_ context.Context, stream chatterbox.ChatterBox_MonitorClient) error {
	// Monitor the connection.
	for {
		msg, err := stream.Recv()
		if err != nil {
			return filterClientError(err)
		}

		switch msg.What {
		case chatterbox.What_CHAT:
			log.Printf("%s: %s", msg.Who, msg.Text)
		case chatterbox.What_JOIN:
			func() {
				mm.mu.Lock()
				defer mm.mu.Unlock()
				mm.members.Add(msg.Who)
			}()
			log.Printf("%s: joined", msg.Who)
		case chatterbox.What_LEAVE:
			func() {
				mm.mu.Lock()
				defer mm.mu.Unlock()
				mm.members.Remove(msg.Who)
			}()
			log.Printf("%s: left", msg.Who)
		default:
			return fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}
