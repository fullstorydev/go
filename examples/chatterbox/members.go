package chatterbox

import (
	"sort"
	"strings"
	"sync"

	"github.com/fullstorydev/go/eventstream"
)

// MembersModel is a basic model object representing the current users in the room.
type MembersModel map[string]struct{}

func (m MembersModel) Add(name string) {
	m[name] = struct{}{}
}

func (m MembersModel) Remove(name string) {
	delete(m, name)
}

func (m MembersModel) String() string {
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return strings.Join(ret, ", ")
}

// MembersList is a synchronized mutable model tracking changes to MembersModel over time.
type MembersList struct {
	mu      sync.RWMutex
	members MembersModel
	es      eventstream.EventStream
}

func NewMembersList() *MembersList {
	return &MembersList{
		members: MembersModel{},
		es:      eventstream.New(),
	}
}

func (m *MembersList) ReadAndSubscribe() ([]string, eventstream.Promise) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ret := make([]string, 0, len(m.members))
	for k := range m.members {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret, m.es.Subscribe()
}

func (m *MembersList) Join(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// apply update, publish event
	m.members.Add(name)
	m.es.Publish(&Event{
		Who:  name,
		What: What_JOIN,
	})
}

func (m *MembersList) Leave(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// apply update, publish event
	m.members.Remove(name)
	m.es.Publish(&Event{
		Who:  name,
		What: What_LEAVE,
	})
}

func (m *MembersList) Chat(name string, text string) {
	m.es.Publish(&Event{
		Who:  name,
		What: What_CHAT,
		Text: text,
	})
}
