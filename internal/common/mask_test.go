package common

import "testing"

func TestShouldMaskKey(t *testing.T) {
	keys := []string{"api_token", "Password", "secret_key", "user"}
	want := []bool{true, true, true, false}
	for i := range keys {
		if got := ShouldMaskKey(keys[i]); got != want[i] {
			t.Fatalf("key %s: got %v want %v", keys[i], got, want[i])
		}
	}
}
