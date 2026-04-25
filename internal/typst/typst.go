package typst

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func EnsureInstalled() error {
	if _, err := exec.LookPath("typst"); err != nil {
		return fmt.Errorf("typst is not installed or not in PATH")
	}
	return nil
}

func Compile(projectRoot, inputFile, outputFile string) error {
	if err := EnsureInstalled(); err != nil {
		return err
	}

	cmd := exec.Command("typst", "compile", "--root", projectRoot, inputFile, outputFile)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, string(output))
	}
	return nil
}

func Watch(projectRoot, inputFile, outputFile string) error {
	if err := EnsureInstalled(); err != nil {
		return err
	}

	cmd := exec.Command("typst", "watch", "--root", projectRoot, inputFile, outputFile)
	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func MainFile(docID string) string {
	return filepath.Join("docs", docID, "main.typ")
}
