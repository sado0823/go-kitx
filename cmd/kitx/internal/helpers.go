package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"golang.org/x/mod/modfile"
)

func KitxHome() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	home := path.Join(dir, ".kitx")
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func KitxHomeWithDir(dir string) string {
	home := path.Join(KitxHome(), dir)
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func GoModulePath(modFile string) (string, error) {
	file, err := os.ReadFile(modFile)
	if err != nil {
		return "", err
	}

	return modfile.ModulePath(file), nil
}

func GitPull(ctx context.Context, url string, path string) error {
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = path
	output, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s \n", output)
	return nil
}

func GitClone(ctx context.Context, url string, branch string, path string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, url, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s \n", output)
	return nil
}

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
