package cli

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractTokenFromQuery(t *testing.T) {
	r := httptest.NewRequest("GET", "/callback?cli_session=abc&token=tok123", nil)
	got, err := extractToken(r, "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tok123" {
		t.Fatalf("got %q, want tok123", got)
	}
}

func TestExtractTokenSessionMismatch(t *testing.T) {
	r := httptest.NewRequest("GET", "/callback?cli_session=evil&token=tok123", nil)
	if _, err := extractToken(r, "abc"); err == nil {
		t.Fatal("session 不匹配时应当拒绝")
	}
}

func TestExtractTokenFromJSONPost(t *testing.T) {
	body := strings.NewReader(`{"cli_session":"abc","token":"tok456"}`)
	r := httptest.NewRequest("POST", "/callback", body)
	r.Header.Set("Content-Type", "application/json")
	got, err := extractToken(r, "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tok456" {
		t.Fatalf("got %q, want tok456", got)
	}
}

func TestExtractTokenFromFormPost(t *testing.T) {
	body := strings.NewReader("cli_session=abc&token=tok789")
	r := httptest.NewRequest("POST", "/callback", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	got, err := extractToken(r, "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tok789" {
		t.Fatalf("got %q, want tok789", got)
	}
}

func TestExtractTokenMissing(t *testing.T) {
	r := httptest.NewRequest("GET", "/callback?cli_session=abc&foo=bar", nil)
	if _, err := extractToken(r, "abc"); err == nil {
		t.Fatal("没有 token 时应当报错")
	}
}
