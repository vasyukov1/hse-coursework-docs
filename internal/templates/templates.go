package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vasyukov1/term-paper/internal/config"
	"github.com/vasyukov1/term-paper/internal/documents"
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

1. Fill in ` + "`term-paper.yaml`" + `.
2. Put the main ` + "`ТЗ`" + ` into the path from ` + "`sources.tz_path`" + `.
3. If you have a team ` + "`ТЗ`" + `, put it into the path from ` + "`sources.team_tz_path`" + `.
4. Put one or more code archives or repository folders into the paths listed in ` + "`sources.code_paths`" + `.
5. Add ` + "`ai.api_key`" + ` or set the environment variable from ` + "`ai.api_key_env`" + `.
6. Run ` + "`term-paper generate`" + ` to scaffold documents and generate AI drafts.
7. Edit generated drafts in ` + "`docs/<doc-id>/drafts`" + `, move the final text into ` + "`sections/`" + `, then run ` + "`term-paper create-pdf`" + `.
`)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme+"\n"), 0o644); err != nil {
		return err
	}

	gitignore := "build/\n.docs-cache/\n"
	return os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0o644)
}
