package plugin

import (
	"bytes"
	"encoding/binary"
	"filippo.io/age/plugin"
	"fmt"
	"strings"
)

const (
	Name = "op"
)

// CreateIdentity creates a new identity.
// Returns the identity and the corresponding recipient.
func CreateIdentity(privateKeyPath string) (*OpIdentity, error) {
	_, err := ReadKeyOp(privateKeyPath)
	if err != nil {
		return nil, err
	}

	identity := NewOpIdentity(privateKeyPath)

	return identity, nil
}

func DecodeIdentity(s string) (*OpIdentity, error) {
	var key OpIdentity
	name, b, err := plugin.ParseIdentity(s)
	if err != nil {
		return nil, err
	}
	if name != Name {
		return nil, fmt.Errorf("invalid hrp")
	}
	r := bytes.NewBuffer(b)
	for _, f := range key.serialize() {
		if err := binary.Read(r, binary.BigEndian, f); err != nil {
			return nil, err
		}
	}

	key.privateKeyPath = strings.TrimPrefix(string(b), "\x01")

	return &key, nil
}
