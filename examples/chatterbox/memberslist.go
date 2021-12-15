package chatterbox

import (
	"sort"
	"sync"

	"github.com/fullstorydev/go/eventstream"
)

type member struct {
	name string
	id   int64
}

type MembersList struct {
	mu      sync.RWMutex
	members map[string]struct{}
	es      eventstream.EventStream
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
	m.members[name] = struct{}{}
	m.es.Publish(&Event{
		Who:  name,
		What: What_JOIN,
	})
}

func (m *MembersList) Leave(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.members, name)
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
