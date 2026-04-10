package datautil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type JWTDecoded struct {
	Header  map[string]any `json:"header"`
	Payload map[string]any `json:"payload"`
}

func DecodeJWT(token string) (JWTDecoded, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return JWTDecoded{}, fmt.Errorf("invalid JWT format")
	}
	headerRaw, err := decodeSegment(parts[0])
	if err != nil {
		return JWTDecoded{}, fmt.Errorf("decode header: %w", err)
	}
	payloadRaw, err := decodeSegment(parts[1])
	if err != nil {
		return JWTDecoded{}, fmt.Errorf("decode payload: %w", err)
	}

	out := JWTDecoded{Header: map[string]any{}, Payload: map[string]any{}}
	if err := json.Unmarshal(headerRaw, &out.Header); err != nil {
		return JWTDecoded{}, fmt.Errorf("parse header JSON: %w", err)
	}
	if err := json.Unmarshal(payloadRaw, &out.Payload); err != nil {
		return JWTDecoded{}, fmt.Errorf("parse payload JSON: %w", err)
	}
	return out, nil
}

func decodeSegment(seg string) ([]byte, error) {
	seg = strings.TrimSpace(seg)
	if m := len(seg) % 4; m != 0 {
		seg += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(seg)
}
