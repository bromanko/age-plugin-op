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
func CreateIdentity(privateKeyPath string) (*OpIdentity, error) {
	_, err := ReadKeyOp(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read private key from 1Password: %v", err)
	}

	identity := ParseIdentity(privateKeyPath)

	return identity, nil
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
