package pending

import (
	"sort"
	"sync"

	"github.com/nicobailon/surf-cli/gohost/internal/host/socketbridge"
)

type RequestKind string

const (
	KindUnknown     RequestKind = "unknown"
	KindToolRequest RequestKind = "tool_request"
	KindStream      RequestKind = "stream_request"
)

type Request struct {
	Session       *socketbridge.Session
	Kind          RequestKind
	OriginalID    any
	HasOriginalID bool
}

type Store struct {
	mu   sync.Mutex
	byID map[int64]Request
}

func NewStore() *Store {
	return &Store{byID: map[int64]Request{}}
}

func (s *Store) Put(id int64, req Request) {
	s.mu.Lock()
	s.byID[id] = req
	s.mu.Unlock()
}

func (s *Store) Pop(id int64) (Request, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, ok := s.byID[id]
	if ok {
		delete(s.byID, id)
	}
	return req, ok
}

func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.byID)
}

func (s *Store) DeleteForSession(session *socketbridge.Session) []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]int64, 0)
	for id, req := range s.byID {
		if req.Session == session {
			ids = append(ids, id)
			delete(s.byID, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
