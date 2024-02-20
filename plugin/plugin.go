package plugin

import (
	"fmt"
)

const (
	Name = "op"
)

// CreateIdentity creates a new identity.
// Returns the identity and the corresponding recipient.
func CreateIdentity(privateKeyPath string) (*OpIdentity, error) {
	_, err := ReadKeyOp(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key from 1Password: %v", err)
	}

	identity := ParseIdentity(privateKeyPath)

	return identity, nil
}
