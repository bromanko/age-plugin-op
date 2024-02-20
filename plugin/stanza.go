package plugin

import (
	"filippo.io/age"
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
