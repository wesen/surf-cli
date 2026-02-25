package nativeio

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	// DefaultMaxFrameSize keeps framing bounded to avoid untrusted oversized payloads.
	DefaultMaxFrameSize = 16 * 1024 * 1024
)

var (
	ErrFrameTooLarge = errors.New("nativeio: frame too large")
	ErrInvalidJSON   = errors.New("nativeio: invalid json")
)

type Decoder struct {
	r       io.Reader
	maxSize uint32
}

func NewDecoder(r io.Reader, maxSize uint32) *Decoder {
	if maxSize == 0 {
		maxSize = DefaultMaxFrameSize
	}
	return &Decoder{r: r, maxSize: maxSize}
}

func (d *Decoder) ReadFrame() ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(d.r, hdr[:]); err != nil {
		return nil, err
	}

	n := binary.LittleEndian.Uint32(hdr[:])
	if n > d.maxSize {
		return nil, fmt.Errorf("%w: %d > %d", ErrFrameTooLarge, n, d.maxSize)
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(d.r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (d *Decoder) ReadJSON(v any) error {
	frame, err := d.ReadFrame()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(frame, v); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return nil
}

func WriteFrame(w io.Writer, payload []byte) error {
	n := len(payload)
	if n > int(DefaultMaxFrameSize) {
		return fmt.Errorf("%w: %d > %d", ErrFrameTooLarge, n, DefaultMaxFrameSize)
	}

	buf := make([]byte, 4+n)
	binary.LittleEndian.PutUint32(buf[:4], uint32(n))
	copy(buf[4:], payload)

	_, err := w.Write(buf)
	return err
}

func WriteJSON(w io.Writer, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return WriteFrame(w, payload)
}
