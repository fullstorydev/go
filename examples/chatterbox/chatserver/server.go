package chatserver

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/fullstorydev/go/examples/chatterbox"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	chatterbox.UnimplementedChatterBoxServer

	model  *ServerMembers
	lastId int64
}

func NewServer() *Server {
	return &Server{
		model:  NewMembersList(),
		lastId: 0,
	}
}

var _ chatterbox.ChatterBoxServer = (*Server)(nil)

func (s *Server) Chat(server chatterbox.ChatterBox_ChatServer) error {
	// Make up a name for this connection.
	id := atomic.AddInt64(&s.lastId, 1)
	name := fmt.Sprintf("User %d", id)

	// Join the memberslist.
	s.model.Join(name)
	log.Printf("%s joined", name)
	defer log.Printf("%s left", name)
	defer s.model.Leave(name)

	// We do not wait on the recv loop to exit; it will exit after we return.
	go func() {
		if err := s.recvLoop(name, server); err != nil {
			log.Printf("%s err: %s", name, err)
		}
	}()

	// Run the send loop in the foreground.
	return s.sendLoop(server)
}

func (s *Server) Monitor(_ *emptypb.Empty, server chatterbox.ChatterBox_MonitorServer) error {
	// Don't join, just monitor.
	return s.sendLoop(server)
}

func (s *Server) recvLoop(name string, server chatterbox.ChatterBox_ChatServer) error {
	for {
		req, err := server.Recv()
		if err != nil {
			return filterServerError(err)
		}

		s.model.Chat(name, req.Text)
		log.Printf("%s: %s", name, req.Text)
	}
}

// commonServerStream intersects ChatterBox_ChatServer and ChatterBox_MonitorServer
type commonServerStream interface {
	grpc.ServerStream
	Send(*chatterbox.Event) error
}

func (s *Server) sendLoop(server commonServerStream) error {
	members, eventPromise := s.model.ReadAndSubscribe()

	// Send the initial members.
	for _, m := range members {
		if err := server.Send(&chatterbox.Event{
			Who:  m,
			What: chatterbox.What_JOIN,
		}); err != nil {
			return filterServerError(err)
		}
	}

	// Signal ready.
	if err := server.Send(&chatterbox.Event{
		What: chatterbox.What_INITIALIZED,
	}); err != nil {
		return filterServerError(err)
	}

	for {
		select {
		case <-server.Context().Done():
			return nil
		case <-eventPromise.Ready():
			evt, nextPromise := eventPromise.Next()
			if nextPromise == nil {
				// end of stream, should never happen
				return nil
			}
			eventPromise = nextPromise

			if err := server.Send(evt.(*chatterbox.Event)); err != nil {
				return filterServerError(err)
			}
		}
	}
}
