package datautil

import "encoding/base64"

func B64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func B64Decode(input string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
