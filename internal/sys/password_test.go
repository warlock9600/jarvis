package sys

import "testing"

func TestGeneratePasswordLength(t *testing.T) {
	pwd, err := GeneratePassword(PasswordOptions{Length: 20, Lower: true, Upper: true, Digits: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pwd) != 20 {
		t.Fatalf("expected length 20, got %d", len(pwd))
	}
}

func TestGeneratePasswordNoAmbiguous(t *testing.T) {
	pwd, err := GeneratePassword(PasswordOptions{Length: 64, Lower: true, Upper: true, Digits: true, NoAmbiguous: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ambiguous := "iloILO01"
	for _, c := range ambiguous {
		for _, p := range pwd {
			if p == c {
				t.Fatalf("password includes ambiguous char %q", c)
			}
		}
	}
}
