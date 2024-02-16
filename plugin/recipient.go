package plugin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"filippo.io/age/plugin"
)

type Recipient struct {
	Pubkey *ed25519.PublicKey
	tag    []byte
}

func (r *Recipient) String() string {
	return EncodeRecipient(r)
}

func NewRecipient(key *ed25519.PublicKey) *Recipient {
	sum := sha256.Sum256(*key)
	return &Recipient{
		Pubkey: key,
		tag:    sum[:4],
	}
}

func EncodeRecipient(recipient *Recipient) string {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.BigEndian, recipient.Pubkey)
	return plugin.EncodeRecipient(Name, b.Bytes())
}
