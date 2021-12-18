package chatclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fullstorydev/go/examples/chatterbox"
)

// RunClient is an example of a gRPC client for a two-way bidi stream.
func RunClient(ctx context.Context, chatInput <-chan string, cl chatterbox.ChatterBoxClient) error {
	mc := &MembersClient{
		cl:        cl,
		chatInput: chatInput,
	}

	if err := mc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	<-ctx.Done() // block forever for this sample
	return nil
}

type MembersClient struct {
	cl        chatterbox.ChatterBoxClient
	chatInput <-chan string

	mu      sync.RWMutex
	members chatterbox.MembersModel
}

// Start this MembersClient. Fetches the initial state synchronously, then background monitors until ctx is cancelled.
func (mc *MembersClient) Start(ctx context.Context) error {
	started := false
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		// Ensure cleanup happens if we do not start successfully.
		if !started {
			cancel()
		}
	}()

	// Synchronously ensure we can fetch an initial model before we return.
	stream, err := mc.startStream(ctx)
	if err != nil {
		return err // failed to fetch the initial state
	}
	started = true

	go func() {
		// Monitor the first stream until it dies.
		if err := mc.monitorStream(ctx, stream); err != nil {
			log.Println(err)
		}
		if ctx.Err() != nil {
			return
		}
		// Run until ctx is cancelled.
		if err := mc.Run(ctx); err != nil {
			log.Println(err)
		}
	}()
	return nil
}

// Run runs this MembersClient in the foreground until ctx is cancelled.
func (mc *MembersClient) Run(ctx context.Context) error {
	const maxBackoff = 16 * time.Second
	backoff := time.Second

	// Loop forever until killed or cancelled.
	for {
		ok, err := func() (bool, error) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			stream, err := mc.startStream(ctx)
			if err != nil {
				return false, err
			}

			return true, mc.monitorStream(ctx, stream)
		}()
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

func (mc *MembersClient) startStream(ctx context.Context) (chatterbox.ChatterBox_ChatClient, error) {
	stream, err := mc.cl.Chat(ctx)
	if err != nil {
		return nil, fmt.Errorf("cl.Chat: %w", err)
	}

	members, err := fetchInitialState(ctx, stream)
	if err != nil {
		return nil, fmt.Errorf("fetchInitialState: %w", err)
	}

	// Successfully fetched initial state.
	log.Printf("Members: %+v", members)
	func() {
		mc.mu.Lock()
		defer mc.mu.Unlock()
		mc.members = members
	}()
	return stream, nil
}

// monitor pulls from the given stream until it closes, updating members model.
func (mc *MembersClient) monitorStream(ctx context.Context, stream chatterbox.ChatterBox_ChatClient) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	// Whenever our stream is up, send any chat inputs to the server.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-mc.chatInput:
				if !ok {
					return
				}
				if err := stream.Send(&chatterbox.Send{
					Text: msg,
				}); err != nil {
					log.Printf("Failed to send: %s", err)
					return
				}
			}
		}
	}()

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
				mc.mu.Lock()
				defer mc.mu.Unlock()
				mc.members.Add(msg.Who)
			}()
			log.Printf("%s: joined", msg.Who)
		case chatterbox.What_LEAVE:
			func() {
				mc.mu.Lock()
				defer mc.mu.Unlock()
				mc.members.Remove(msg.Who)
			}()
			log.Printf("%s: left", msg.Who)
		default:
			return fmt.Errorf("unexpected type: %s", msg.What)
		}
	}
}
