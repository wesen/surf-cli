package socketbridge

import (
	"encoding/json"
	"net"
	"sync"
)

type Session struct {
	conn    net.Conn
	writeMu sync.Mutex
}

func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Conn() net.Conn {
	return s.conn
}

func (s *Session) WriteLine(line []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	buf := append([]byte(nil), line...)
	if len(buf) == 0 || buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}
	_, err := s.conn.Write(buf)
	return err
}

func (s *Session) WriteJSONLine(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.WriteLine(b)
}

func (s *Session) Close() error {
	return s.conn.Close()
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[*Session]struct{}
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: map[*Session]struct{}{}}
}

func (m *SessionManager) Add(conn net.Conn) *Session {
	s := NewSession(conn)
	m.mu.Lock()
	m.sessions[s] = struct{}{}
	m.mu.Unlock()
	return s
}

func (m *SessionManager) Remove(s *Session) {
	if s == nil {
		return
	}
	m.mu.Lock()
	delete(m.sessions, s)
	m.mu.Unlock()
}

func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

func (m *SessionManager) Snapshot() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessions))
	for s := range m.sessions {
		out = append(out, s)
	}
	return out
}

func (m *SessionManager) NotifyExtensionDisconnected(message string) {
	payload := map[string]any{
		"type":    "extension_disconnected",
		"message": message,
	}
	for _, s := range m.Snapshot() {
		_ = s.WriteJSONLine(payload)
		_ = s.Close()
		m.Remove(s)
	}
}

func (m *SessionManager) CloseAll() {
	for _, s := range m.Snapshot() {
		_ = s.Close()
		m.Remove(s)
	}
}
