package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vasyukov1/hse-coursework-docs/internal/config"
	"github.com/vasyukov1/hse-coursework-docs/internal/documents"
)

//go:embed all:assets/project/**
var projectFS embed.FS

func WriteProject(root string, cfg config.Config, selected []documents.Spec) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := config.Save(root, cfg); err != nil {
		return err
	}
	if err := copyDir("assets/project/shared", filepath.Join(root, "shared")); err != nil {
		return err
	}
	for _, spec := range selected {
		if err := copyDir(filepath.Join("assets/project/docs", spec.Folder), filepath.Join(root, "docs", spec.Folder)); err != nil {
			return err
		}
	}

	return writeRootFiles(root)
}

func copyDir(src, dst string) error {
	return fs.WalkDir(projectFS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return os.MkdirAll(dst, 0o755)
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		}
		data, err := projectFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func writeRootFiles(root string) error {
	readme := strings.TrimSpace(`
# term-paper project

1. Заполните ` + "`term-paper.yaml`" + `.
2. Положите основное ТЗ в ` + "`input/tz`" + `.
3. Если нужен командный ПМИ, положите командное ТЗ в ` + "`input/tz-team`" + `.
4. Положите архивы кода или каталоги репозиториев в ` + "`input/code`" + `.
5. Допишите заметки в ` + "`input/notes.txt`" + `.
6. Запустите ` + "`term-paper generate-doc --doc pz`" + ` или другую нужную команду.
7. Проверьте ` + "`docs/<doc>/drafts`" + ` или ` + "`docs/<doc>/sections`" + `.
8. Соберите PDF командой ` + "`term-paper create-pdf --doc <doc>`" + `.
`)
	if err := writeIfMissing(filepath.Join(root, "README.md"), []byte(readme+"\n")); err != nil {
		return err
	}

	gitignore := "build/\n.docs-cache/\n"
	return writeIfMissing(filepath.Join(root, ".gitignore"), []byte(gitignore))
}

func writeIfMissing(path string, data []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, data, 0o644)
}
