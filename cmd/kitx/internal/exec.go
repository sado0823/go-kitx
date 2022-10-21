package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GoInstall(paths ...string) error {
	for _, path := range paths {
		if !strings.Contains(path, "@") {
			path += "@latest"
		}
		fmt.Printf("go install %s \n", path)
		cmd := exec.Command("go", "install", path)
		cmd.Stderr = os.Stdout
		cmd.Stdout = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
