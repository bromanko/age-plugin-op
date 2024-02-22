package plugin

import (
	"fmt"
	"os/exec"
)

func ReadKeyOp(privateKeyPath string) ([]byte, error) {
	cmd := exec.Command("op", "read", privateKeyPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not read private key from 1Password: %v", err)
	}
	return output, nil
}
