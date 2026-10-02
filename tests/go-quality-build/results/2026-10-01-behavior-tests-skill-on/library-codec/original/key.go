package keycodec

import (
	"errors"
	"strings"
)

var ErrInvalidKey = errors.New("invalid key")

type Key struct {
	Region string
	Name   string
}

func Encode(k Key) (string, error) {
	if k.Region == "" || k.Name == "" || strings.ContainsAny(k.Region+k.Name, "/%") {
		return "", ErrInvalidKey
	}
	return k.Region + "/" + k.Name, nil
}

func Decode(wire string) (Key, error) {
	p := strings.Split(wire, "/")
	if len(p) != 2 || p[0] == "" || p[1] == "" || strings.Contains(wire, "%") {
		return Key{}, ErrInvalidKey
	}
	return Key{Region: p[0], Name: p[1]}, nil
}
