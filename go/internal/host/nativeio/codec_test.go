package nativeio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWriteAndReadJSONRoundTrip(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	in := map[string]any{"type": "HOST_READY", "ok": true}

	if err := WriteJSON(&b, in); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	var out map[string]any
	dec := NewDecoder(&b, DefaultMaxFrameSize)
	if err := dec.ReadJSON(&out); err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}

	if out["type"] != "HOST_READY" {
		t.Fatalf("unexpected type: %#v", out["type"])
	}
	if out["ok"] != true {
		t.Fatalf("unexpected ok: %#v", out["ok"])
	}
}

func TestDecoderRejectsFrameTooLarge(t *testing.T) {
	t.Parallel()

	payload := []byte("12345")
	var b bytes.Buffer
	if err := WriteFrame(&b, payload); err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}

	dec := NewDecoder(&b, 4)
	_, err := dec.ReadFrame()
	if err == nil || !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected ErrFrameTooLarge, got %v", err)
	}
}

func TestDecoderInvalidJSON(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	if err := WriteFrame(&b, []byte("not-json")); err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}

	dec := NewDecoder(&b, DefaultMaxFrameSize)
	var v map[string]any
	err := dec.ReadJSON(&v)
	if err == nil || !errors.Is(err, ErrInvalidJSON) {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestDecoderUnexpectedEOF(t *testing.T) {
	t.Parallel()

	// Header says length 10, but only 3 payload bytes follow.
	data := []byte{10, 0, 0, 0, 'a', 'b', 'c'}
	dec := NewDecoder(bytes.NewReader(data), DefaultMaxFrameSize)

	_, err := dec.ReadFrame()
	if err == nil || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestWriteFrameTooLarge(t *testing.T) {
	t.Parallel()

	big := strings.Repeat("x", DefaultMaxFrameSize+1)
	var b bytes.Buffer
	err := WriteFrame(&b, []byte(big))
	if err == nil || !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected ErrFrameTooLarge, got %v", err)
	}
}
