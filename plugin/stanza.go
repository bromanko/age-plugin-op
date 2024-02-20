package plugin

import (
	"filippo.io/age"
	"io"
	"strconv"
)

func NewIndexedErrorStanza(kind string, index int, err error) *age.Stanza {
	return &age.Stanza{
		Type: "error",
		Args: []string{kind, strconv.Itoa(index)},
		Body: []byte(err.Error()),
	}
}

func NewInternalErrorStanza(err error) *age.Stanza {
	return &age.Stanza{
		Type: "error",
		Args: []string{"internal"},
		Body: []byte(err.Error()),
	}
}

var stanzaPrefix = []byte("->")

func MarshalStanza(s *age.Stanza, w io.Writer) error {
	if _, err := w.Write(stanzaPrefix); err != nil {
		return err
	}
	for _, a := range append([]string{s.Type}, s.Args...) {
		if _, err := io.WriteString(w, " "+a); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	ww := NewWrappedBase64Encoder(b64, w)
	if _, err := ww.Write(s.Body); err != nil {
		return err
	}
	if err := ww.Close(); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}
