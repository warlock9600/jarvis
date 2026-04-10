package sys

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

const (
	lowerChars = "abcdefghjkmnpqrstuvwxyz"
	upperChars = "ABCDEFGHJKMNPQRSTUVWXYZ"
	digitChars = "23456789"
	symbols    = "!@#$%^&*()-_=+[]{}<>?"
)

type PasswordOptions struct {
	Length      int
	Lower       bool
	Upper       bool
	Digits      bool
	Symbols     bool
	NoAmbiguous bool
	Profile     string
}

func ApplyProfile(opts PasswordOptions) PasswordOptions {
	switch strings.ToLower(opts.Profile) {
	case "infra":
		opts.Length = maxInt(opts.Length, 24)
		opts.Lower, opts.Upper, opts.Digits, opts.Symbols = true, true, true, true
	case "human":
		opts.Length = maxInt(opts.Length, 16)
		opts.Lower, opts.Upper, opts.Digits, opts.Symbols = true, true, true, false
		opts.NoAmbiguous = true
	case "strict":
		opts.Length = maxInt(opts.Length, 32)
		opts.Lower, opts.Upper, opts.Digits, opts.Symbols = true, true, true, true
		opts.NoAmbiguous = true
	}
	return opts
}

func GeneratePassword(opts PasswordOptions) (string, error) {
	opts = ApplyProfile(opts)
	if opts.Length <= 0 {
		return "", errors.New("password length must be greater than 0")
	}
	if !opts.Lower && !opts.Upper && !opts.Digits && !opts.Symbols {
		return "", errors.New("at least one character group must be enabled")
	}

	pool := ""
	if opts.Lower {
		pool += lowerChars
	}
	if opts.Upper {
		pool += upperChars
	}
	if opts.Digits {
		pool += digitChars
	}
	if opts.Symbols {
		pool += symbols
	}

	if !opts.NoAmbiguous {
		if opts.Lower {
			pool += "ilo"
		}
		if opts.Upper {
			pool += "ILO"
		}
		if opts.Digits {
			pool += "01"
		}
	}

	buf := make([]byte, opts.Length)
	for i := 0; i < opts.Length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
		if err != nil {
			return "", err
		}
		buf[i] = pool[idx.Int64()]
	}
	return string(buf), nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
