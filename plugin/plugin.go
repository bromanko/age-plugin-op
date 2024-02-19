package plugin

import (
	"filippo.io/age"
	"filippo.io/age/agessh"
	"fmt"
	"os"
)

const (
	Name = "op"
)

// CreateIdentity creates a new identity.
// Returns the identity and the corresponding recipient.
func CreateIdentity(privateKeyPath string) (*age.Identity, *age.Recipient, error) {
	// TODO - replace this with load from 1Password
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read private key: %v", err)
	}

	identity, err := agessh.ParseIdentity(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse identity from private key: %v", err)
	}

	var recipient age.Recipient
	switch id := identity.(type) {
	case *agessh.Ed25519Identity:
		recipient = id.Recipient()
	case *agessh.RSAIdentity:
		recipient = id.Recipient()
	}

	return &identity, &recipient, nil
}
