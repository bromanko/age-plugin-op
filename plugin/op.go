package plugin

import (
	"os"
)

func ReadKeyOp(privateKeyPath string) ([]byte, error) {
	// TODO - replace this with load from 1Password
	return os.ReadFile(privateKeyPath)
}
