package plugin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"filippo.io/age"
	"filippo.io/age/plugin"
	"filippo.io/edwards25519"
	"fmt"
	"golang.org/x/crypto/ssh"
)

func encodeOpSSHDetails(op *OpSSHDetails) string {
	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, []byte(op.opPath))
	if err != nil {
		Log.Println("failed to encode recipient: %v", err)
	}
	return plugin.EncodeRecipient(Name, b.Bytes())
}

type OpSSHDetails struct {
	opPath string
	tag    []byte
}

type RSARecipient struct {
	op     *OpSSHDetails
	sshKey ssh.PublicKey
	pubKey *rsa.PublicKey
}

func (r *RSARecipient) String() string {
	return encodeOpSSHDetails(r.op)
}

func NewRSARecipient(pk ssh.PublicKey) (*RSARecipient, error) {
	if pk.Type() != "ssh-rsa" {
		return nil, errors.New("SSH public key is not an RSA key")
	}
	r := &RSARecipient{
		sshKey: pk,
	}

	if pk, ok := pk.(ssh.CryptoPublicKey); ok {
		if pk, ok := pk.CryptoPublicKey().(*rsa.PublicKey); ok {
			r.pubKey = pk
		} else {
			return nil, errors.New("unexpected public key type")
		}
	} else {
		return nil, errors.New("pk does not implement ssh.CryptoPublicKey")
	}
	if r.pubKey.Size() < 2048/8 {
		return nil, errors.New("RSA key size is too small")
	}
	return r, nil
}

type Ed25519Recipient struct {
	op             *OpSSHDetails
	sshKey         ssh.PublicKey
	theirPublicKey []byte
}

func (r *Ed25519Recipient) String() string {
	return encodeOpSSHDetails(r.op)
}

func ed25519PublicKeyToCurve25519(pk ed25519.PublicKey) ([]byte, error) {
	// See https://blog.filippo.io/using-ed25519-keys-for-encryption and
	// https://pkg.go.dev/filippo.io/edwards25519#Point.BytesMontgomery.
	p, err := new(edwards25519.Point).SetBytes(pk)
	if err != nil {
		return nil, err
	}
	return p.BytesMontgomery(), nil
}

func NewEd25519Recipient(pk ssh.PublicKey) (*Ed25519Recipient, error) {
	if pk.Type() != "ssh-ed25519" {
		return nil, errors.New("SSH public key is not an Ed25519 key")
	}

	cpk, ok := pk.(ssh.CryptoPublicKey)
	if !ok {
		return nil, errors.New("pk does not implement ssh.CryptoPublicKey")
	}
	epk, ok := cpk.CryptoPublicKey().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("unexpected public key type")
	}
	mpk, err := ed25519PublicKeyToCurve25519(epk)
	if err != nil {
		return nil, fmt.Errorf("invalid Ed25519 public key: %v", err)
	}

	return &Ed25519Recipient{
		sshKey:         pk,
		theirPublicKey: mpk,
	}, nil
}

// ###############################################################33

type OpRecipient struct {
	Recipient      *age.Recipient
	privateKeyPath string
	tag            []byte
}

// Tag returns the 4 first bytes of the path to the key
// this is used to find the correct identity in a stanza
func (r *OpRecipient) Tag() []byte {
	return r.tag
}

func (r *OpRecipient) String() string {
	return EncodeRecipient(r)
}

func NewRecipient(opPath string, r *age.Recipient) *OpRecipient {
	sum := sha256.Sum256([]byte(opPath))
	return &OpRecipient{
		Recipient:      r,
		privateKeyPath: opPath,
		tag:            sum[:4],
	}
}

func EncodeRecipient(r *OpRecipient) string {
	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, []byte(r.privateKeyPath))
	if err != nil {
		Log.Println("failed to encode recipient: %v", err)
	}
	return plugin.EncodeRecipient(Name, b.Bytes())
}

func DecodeOpKeyPath(s string) (string, error) {
	name, b, err := plugin.ParseRecipient(s)
	if err != nil {
		return "", fmt.Errorf("failed to decode recipient: %v", err)
	}
	if name != Name {
		return "", fmt.Errorf("invalid plugin for type %s", name)
	}

	return string(b), nil
}
