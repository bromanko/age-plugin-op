package plugin

import (
	"fmt"
	"io"
	"strings"
	"time"
)

//type Identity struct {
//	Version uint8
//	Private *ed25519.PrivateKey
//	Public  *ed25519.PublicKey
//}
//
//func (i *Identity) Serialize() []any {
//	return []interface{}{
//		&i.Version,
//	}
//}
//
//func (i *Identity) Recipient() *Recipient {
//	return NewRecipient(i.Public)
//}
//
//func EncodeIdentity(i *Identity) string {
//	var b bytes.Buffer
//	for _, v := range i.Serialize() {
//		_ = binary.Write(&b, binary.BigEndian, v)
//	}
//
//	var pub []byte
//	pub = append(pub, *i.Public...)
//	pub = append(pub, *i.Private...)
//	b.Write(pub)
//
//	return plugin.EncodeIdentity(Name, b.Bytes())
//}

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

func MarshalIdentity(recipient *OpRecipient, w io.Writer) error {
	Marshal(w)
	_, _ = fmt.Fprintf(w, "# Recipient: %s\n", recipient)
	//_, _ = fmt.Fprintf(w, "\n%s\n", EncodeIdentity(i))
	return nil
}
