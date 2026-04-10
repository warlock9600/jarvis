package catutil

import "testing"

func TestIsProbablyBinary(t *testing.T) {
	if !isProbablyBinary([]byte{0x00, 0x01, 0x02}) {
		t.Fatal("expected binary detection for null bytes")
	}
	if isProbablyBinary([]byte("hello\nworld\n")) {
		t.Fatal("expected plain text to be non-binary")
	}
}

func TestDetectLexer(t *testing.T) {
	got := detectLexer("sample.json", "{\"ok\":true}")
	if got == "" || got == "plaintext" {
		t.Fatalf("expected lexer for json, got %q", got)
	}
}
