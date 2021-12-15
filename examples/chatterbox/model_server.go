package chatterbox

import (
	"sort"
	"sync"

	"github.com/fullstorydev/go/eventstream"
)

// ServerMembers is a server-side synchronized mutable model tracking changes to MembersModel over time.
type ServerMembers struct {
	mu      sync.RWMutex
	members MembersModel
	es      eventstream.EventStream
}

func NewMembersList() *ServerMembers {
	return &ServerMembers{
		members: MembersModel{},
		es:      eventstream.New(),
	}
}

func (m *ServerMembers) ReadAndSubscribe() ([]string, eventstream.Promise) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ret := make([]string, 0, len(m.members))
	for k := range m.members {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret, m.es.Subscribe()
}

func (m *ServerMembers) Join(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// apply update, publish event
	m.members.Add(name)
	m.es.Publish(&Event{
		Who:  name,
		What: What_JOIN,
	})
}

func (m *ServerMembers) Leave(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// apply update, publish event
	m.members.Remove(name)
	m.es.Publish(&Event{
		Who:  name,
		What: What_LEAVE,
	})
}

func (m *ServerMembers) Chat(name string, text string) {
	m.es.Publish(&Event{
		Who:  name,
		What: What_CHAT,
		Text: text,
	})
}
