package plugin

import (
	"bytes"
	"encoding/binary"
	"filippo.io/age"
	"filippo.io/age/plugin"
	"fmt"
	"io"
	"strings"
	"time"
)

const version = 1

type OpIdentity struct {
	Version        uint8
	privateKeyPath string
}

var _ age.Identity = &OpIdentity{}

func (i *OpIdentity) serialize() []any {
	return []interface{}{
		&i.Version,
	}
}

func (i *OpIdentity) Unwrap(_ []*age.Stanza) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func NewOpIdentity(privateKeyPath string) *OpIdentity {
	i := &OpIdentity{
		Version:        version,
		privateKeyPath: privateKeyPath,
	}
	return i
}

func (i *OpIdentity) Recipient() *OpRecipient {
	return &OpRecipient{
		privateKeyPath: i.privateKeyPath,
	}
}

func ParseIdentity(privateKeyPath string) *OpIdentity {
	return NewOpIdentity(privateKeyPath)
}

func encodeIdentity(i *OpIdentity) string {
	var b bytes.Buffer
	for _, v := range i.serialize() {
		_ = binary.Write(&b, binary.BigEndian, v)
	}

	err := binary.Write(&b, binary.BigEndian, []byte(i.privateKeyPath))
	if err != nil {
		Log.Printf("failed to encode identity: %v", err)
	}

	return plugin.EncodeIdentity(Name, b.Bytes())
}

var (
	marshalTemplate = `
# Created: %s
`
)

func Marshal(w io.Writer) {
	s := fmt.Sprintf(marshalTemplate, time.Now())
	s = strings.TrimSpace(s)
	_, _ = fmt.Fprintf(w, "%s\n", s)
}

func MarshalIdentity(i *OpIdentity, w io.Writer) error {
	Marshal(w)
	_, _ = fmt.Fprintf(w, "# Recipient: %s\n", i.Recipient())
	_, _ = fmt.Fprintf(w, "\n%s\n", encodeIdentity(i))
	return nil
}
