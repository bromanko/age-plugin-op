package plugin

import (
	"filippo.io/age"
	"filippo.io/age/agessh"
	"fmt"
)

const (
	Name = "op"
)

// CreateIdentity creates a new identity.
// Returns the identity and the corresponding recipient.
func CreateIdentity(privateKeyPath string) (*age.Identity, *OpRecipient, error) {
	privateKey, err := ReadKeyOp(privateKeyPath)
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
	r := NewRecipient(privateKeyPath, &recipient)

	return &identity, r, nil
}

func EncryptFileKey(fileKey []byte, i *age.Identity) ([]*age.Stanza, error) {

	switch sshR := (*i).(type) {
	case *agessh.Ed25519Identity:
		return sshR.Recipient().Wrap(fileKey)
	case *agessh.RSAIdentity:
		return sshR.Recipient().Wrap(fileKey)
	default:
		return nil, fmt.Errorf("unsupported recipient type: %T", sshR)
	}
}
