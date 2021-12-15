package chatterbox

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/fullstorydev/go/eventstream"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	UnimplementedChatterBoxServer

	model  MembersList
	nextId int64
}

func NewServer() *Server {
	return &Server{
		model: MembersList{
			mu:      sync.RWMutex{},
			members: map[string]struct{}{},
			es:      eventstream.New(),
		},
	}
}

var _ ChatterBoxServer = (*Server)(nil)

func (s *Server) Chat(server ChatterBox_ChatServer) error {
	// Make up a name for this connection.
	id := atomic.AddInt64(&s.nextId, 1)
	name := fmt.Sprintf("User %d", id)

	// Join the memberslist.
	s.model.Join(name)
	log.Printf("%s joined", name)
	defer log.Printf("%s left", name)
	defer s.model.Leave(name)

	// We do not wait on this loop to exit; it will exit after we return.
	go func() {
		if err := s.recvLoop(name, server); err != nil {
			log.Printf("%s err: %s", name, err)
		}
	}()
	return s.sendLoop(server)
}

func (s *Server) Listen(empty *emptypb.Empty, server ChatterBox_ListenServer) error {
	// Don't join, just listen.
	return s.sendLoop(server)
}

type commonServer interface {
	grpc.ServerStream
	Send(*Event) error
}

func (s *Server) recvLoop(name string, server ChatterBox_ChatServer) error {
	for {
		req, err := server.Recv()
		if err != nil {
			return filterErr(err)
		}

		s.model.Chat(name, req.Text)
		log.Printf("%s: %s", name, req.Text)
	}
}

func (s *Server) sendLoop(server commonServer) error {
	members, eventPromise := s.model.ReadAndSubscribe()

	// Send the initial members.
	for _, m := range members {
		if err := server.Send(&Event{
			Who:  m,
			What: What_JOIN,
		}); err != nil {
			return filterErr(err)
		}
	}

	// Signal ready.
	if err := server.Send(&Event{
		What: What_INITIALIZED,
	}); err != nil {
		return filterErr(err)
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

			if err := server.Send(evt.(*Event)); err != nil {
				return filterErr(err)
			}
		}
	}
}
