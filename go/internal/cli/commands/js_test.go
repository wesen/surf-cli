package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildJSArgsInline(t *testing.T) {
	args, err := buildJSArgs(&JSSettings{Code: "return document.title"})
	if err != nil {
		t.Fatalf("buildJSArgs returned error: %v", err)
	}
	if got := args["code"]; got != "return document.title" {
		t.Fatalf("unexpected code: %#v", got)
	}
}

func TestBuildJSArgsFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.js")
	if err := os.WriteFile(path, []byte("return document.title\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	args, err := buildJSArgs(&JSSettings{File: path})
	if err != nil {
		t.Fatalf("buildJSArgs returned error: %v", err)
	}
	if got := args["code"]; got != "return document.title\n" {
		t.Fatalf("unexpected file-backed code: %#v", got)
	}
}

func TestBuildJSArgsRejectsBothInlineAndFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "script.js")
	if err := os.WriteFile(path, []byte("return 1"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := buildJSArgs(&JSSettings{Code: "return 2", File: path})
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestBuildJSArgsRequiresInput(t *testing.T) {
	_, err := buildJSArgs(&JSSettings{})
	if err == nil {
		t.Fatalf("expected missing input error")
	}
}
