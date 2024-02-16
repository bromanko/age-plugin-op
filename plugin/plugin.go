package plugin

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
)

const (
	Name = "op"
)

// CreateIdentity Creates a new identity.
// Returns the identity and the corresponding recipient.
func CreateIdentity() (*Identity, *Recipient, error) {
	// TODO - replace this with load from 1Password
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating ED25519 key: %v", err)
	}

	identity := &Identity{
		Version: 1,
		Private: &privateKey,
		Public:  &publicKey,
	}

	recipient := identity.Recipient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed getting recipient: %v", err)
	}
	return identity, recipient, nil
}
