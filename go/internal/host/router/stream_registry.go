package router

import (
	"sort"
	"sync"

	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

type StreamRegistry struct {
	mu       sync.Mutex
	byStream map[int64]*socketbridge.Session
}

func NewStreamRegistry() *StreamRegistry {
	return &StreamRegistry{byStream: map[int64]*socketbridge.Session{}}
}

func (r *StreamRegistry) Set(streamID int64, session *socketbridge.Session) {
	r.mu.Lock()
	r.byStream[streamID] = session
	r.mu.Unlock()
}

func (r *StreamRegistry) Get(streamID int64) (*socketbridge.Session, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byStream[streamID]
	return s, ok
}

func (r *StreamRegistry) Delete(streamID int64) {
	r.mu.Lock()
	delete(r.byStream, streamID)
	r.mu.Unlock()
}

func (r *StreamRegistry) DeleteBySession(session *socketbridge.Session) []int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]int64, 0)
	for streamID, s := range r.byStream {
		if s == session {
			ids = append(ids, streamID)
			delete(r.byStream, streamID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r *StreamRegistry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byStream)
}
