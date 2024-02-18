package plugin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"filippo.io/age/plugin"
	"fmt"
)

type Recipient struct {
	Pubkey *ed25519.PublicKey
	tag    []byte
}

// Tag returns the 4 first bytes of a sha256 sum of the key
// this is used to find the correct identity in a stanza
func (r *Recipient) Tag() []byte {
	return r.tag
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

func DecodeRecipient(s string) (*Recipient, error) {
	name, b, err := plugin.ParseRecipient(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode recipient: %v", err)
	}
	if name != Name {
		return nil, fmt.Errorf("invalid plugin for type %s", name)
	}

	key := ed25519.PublicKey(b)

	return NewRecipient(&key), nil
}
