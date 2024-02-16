package plugin

import (
	"crypto/ed25519"
	"crypto/sha256"
)

type Recipient struct {
	Pubkey *ed25519.PublicKey
	tag    []byte
}

func NewRecipient(key *ed25519.PublicKey) *Recipient {
	sum := sha256.Sum256(*key)
	return &Recipient{
		Pubkey: key,
		tag:    sum[:4],
	}
}
