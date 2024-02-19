package plugin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"filippo.io/age"
	"filippo.io/age/plugin"
	"fmt"
)

type OpRecipient struct {
	Recipient *age.Recipient
	opPath    string
	tag       []byte
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
		Recipient: r,
		opPath:    opPath,
		tag:       sum[:4],
	}
}

func EncodeRecipient(r *OpRecipient) string {
	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, []byte(r.opPath))
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
