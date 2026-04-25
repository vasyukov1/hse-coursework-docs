package typst

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var includePattern = regexp.MustCompile(`^\s*#include\s+"([^"]+)"`)

func BundleSource(inputPath, entry string) (string, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return bundleFile(inputPath, map[string]bool{})
	}

	if entry != "" {
		entryPath := filepath.Join(inputPath, entry)
		if _, err := os.Stat(entryPath); err == nil {
			return bundleFile(entryPath, map[string]bool{})
		}
	}

	for _, candidate := range []string{"main.typ", "body.typ"} {
		candidatePath := filepath.Join(inputPath, candidate)
		if _, err := os.Stat(candidatePath); err == nil {
			return bundleFile(candidatePath, map[string]bool{})
		}
	}

	return concatTypstFiles(inputPath)
}

func bundleFile(path string, seen map[string]bool) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if seen[absPath] {
		return "", nil
	}
	seen[absPath] = true

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	var parts []string
	parts = append(parts, "// BEGIN "+filepath.Base(absPath))

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		matches := includePattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			includePath := filepath.Join(filepath.Dir(absPath), matches[1])
			inlined, err := bundleFile(includePath, seen)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(inlined) != "" {
				parts = append(parts, inlined)
			}
			continue
		}
		parts = append(parts, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	parts = append(parts, "// END "+filepath.Base(absPath))
	return strings.Join(parts, "\n"), nil
}

func concatTypstFiles(root string) (string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".typ" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "", fmt.Errorf("no .typ files found in %s", root)
	}

	var parts []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			rel = file
		}
		parts = append(parts, "// BEGIN "+rel, string(data), "// END "+rel)
	}
	return strings.Join(parts, "\n\n"), nil
}
