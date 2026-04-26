package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vasyukov1/hse-coursework-docs/internal/config"
)

func TestGenerateDocSubset(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pmi",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "pmi", "main.typ")); err != nil {
		t.Fatalf("expected pmi doc to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "tz")); !os.IsNotExist(err) {
		t.Fatalf("expected tz doc to be absent, got err=%v", err)
	}
}

func TestGenerateDocWritesSectionFiles(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pz",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	projectDir := filepath.Join(root, "input", "code", "sample-project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "README.md"), []byte("# Sample project\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "== Сгенерированный текст\n\nTODO: Уточнить детали проекта.\n",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.AI.APIKey = "test-key"
	cfg.AI.Provider = "openrouter"
	cfg.AI.BaseURL = server.URL
	cfg.AI.StyleExamples = map[string][]string{}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	restore := chdir(t, root)
	defer restore()

	originalSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "02-requirements.typ"))
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pz",
		Output: root,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	currentSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "02-requirements.typ"))
	if err != nil {
		t.Fatalf("expected generated section: %v", err)
	}
	if !strings.Contains(string(currentSection), "Сгенерированный текст") {
		t.Fatalf("unexpected section content: %s", string(currentSection))
	}

	if string(currentSection) == string(originalSection) {
		t.Fatalf("expected working section file to be updated by default")
	}
}

func TestGenerateDocCanWriteDraftFiles(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pz",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "== Черновой текст\n\nTODO: Уточнить детали проекта.\n",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.AI.APIKey = "test-key"
	cfg.AI.Provider = "openrouter"
	cfg.AI.BaseURL = server.URL
	cfg.AI.StyleExamples = map[string][]string{}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	restore := chdir(t, root)
	defer restore()

	originalSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "02-requirements.typ"))
	if err != nil {
		t.Fatal(err)
	}

	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pz",
		Output: root,
		Draft:  true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	draft, err := os.ReadFile(filepath.Join(root, "docs", "pz", "drafts", "02-requirements.typ"))
	if err != nil {
		t.Fatalf("expected generated draft: %v", err)
	}
	if !strings.Contains(string(draft), "Черновой текст") {
		t.Fatalf("unexpected draft content: %s", string(draft))
	}

	currentSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "02-requirements.typ"))
	if err != nil {
		t.Fatal(err)
	}
	if string(currentSection) != string(originalSection) {
		t.Fatalf("expected working section file to stay unchanged in draft mode")
	}
}

func TestImproveDocWritesImprovedCopy(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "ro",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "= ВЫПОЛНЕНИЕ ПРОГРАММЫ\n\n#h(2em) Уточненный текст.\n",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.AI.APIKey = "test-key"
	cfg.AI.Provider = "openrouter"
	cfg.AI.BaseURL = server.URL
	cfg.AI.StyleExamples = map[string][]string{}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	restore := chdir(t, root)
	defer restore()

	if err := ImproveDoc(ImproveDocOptions{
		File:   "docs/ro/sections/03-run.typ",
		Prompt: "сделай текст подробнее",
	}); err != nil {
		t.Fatalf("ImproveDoc() error = %v", err)
	}

	improvedPath := filepath.Join(root, "docs", "ro", "sections", "03-run.typ.improved.typ")
	data, err := os.ReadFile(improvedPath)
	if err != nil {
		t.Fatalf("expected improved file: %v", err)
	}
	if !strings.Contains(string(data), "Уточненный текст") {
		t.Fatalf("unexpected improved content: %s", string(data))
	}
}

func TestDoctorAndCreatePDF(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	if err := GenerateDoc(GenerateDocOptions{
		Doc:    "pmi",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("GenerateDoc() error = %v", err)
	}

	if _, err := execLookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}

	restore := chdir(t, root)
	defer restore()

	if _, err := Doctor(); err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}

	if err := CreatePDF(CreatePDFOptions{
		Doc:    "all",
		Output: "build",
	}); err != nil {
		t.Fatalf("CreatePDF() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "build", "pmi.pdf")); err != nil {
		t.Fatalf("expected pmi.pdf: %v", err)
	}
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	}
}

func execLookPath(bin string) (string, error) {
	return exec.LookPath(bin)
}

func TestBundleTypst(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.typ"), []byte("#include \"body.typ\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "body.typ"), []byte("= Root\n#include \"sections/01.typ\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sections", "01.typ"), []byte("== Section\nText.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(root, "bundled.typ")
	if err := BundleTypst(BundleTypstOptions{
		Input:  root,
		Output: out,
		Entry:  "main.typ",
	}); err != nil {
		t.Fatalf("BundleTypst() error = %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "= Root") || !strings.Contains(text, "== Section") {
		t.Fatalf("unexpected bundled content: %s", text)
	}
}

func initProject(t *testing.T, root string) {
	t.Helper()
	if err := Init(InitOptions{
		Output: root,
	}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() after init error = %v", err)
	}
	cfg.AI.APIKey = "test-key"
	cfg.AI.Provider = "openrouter"
	cfg.AI.BaseURL = "https://example.invalid"
	cfg.AI.StyleExamples = map[string][]string{}
	cfg.Sources.TZDir = "./input/tz"
	cfg.Sources.TeamTZDir = "./input/tz-team"
	cfg.Sources.CodeDir = "./input/code"
	cfg.Sources.NotesFile = "./input/notes.txt"
	cfg.Sources.SourcePriority = "balanced"
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() after init error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "input", "tz", "main.typ"), []byte("= ТЗ\n\nТекст технического задания.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "input", "tz-team", "main.typ"), []byte("= Командное ТЗ\n\nТекст командного ТЗ.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "input", "notes.txt"), []byte("TODO: вставить скрин главного экрана.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
