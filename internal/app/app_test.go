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

	"github.com/vasyukov1/term-paper/internal/config"
)

func TestGenerateSubset(t *testing.T) {
	root := t.TempDir()
	initProject(t, root, "single")

	if err := Generate(GenerateOptions{
		Doc:    "pmi",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "pmi", "main.typ")); err != nil {
		t.Fatalf("expected pmi doc to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "tz")); !os.IsNotExist(err) {
		t.Fatalf("expected tz doc to be absent, got err=%v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.Docs["pmi"].Enabled {
		t.Fatalf("expected pmi to be enabled")
	}
	if cfg.Docs["pmi-team"].Enabled {
		t.Fatalf("expected pmi-team to stay disabled in single mode")
	}
}

func TestGenerateTeamProjectEnablesTeamPMI(t *testing.T) {
	root := t.TempDir()
	initProject(t, root, "team")

	if err := Generate(GenerateOptions{
		Doc:    "all",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Participants) < 2 {
		t.Fatalf("expected team mode to create multiple participants")
	}
	if !cfg.Docs["pmi-team"].Enabled {
		t.Fatalf("expected pmi-team to be enabled")
	}
}

func TestDoctorAndCreatePDF(t *testing.T) {
	root := t.TempDir()
	initProject(t, root, "single")
	if err := Generate(GenerateOptions{
		Doc:    "all",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
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
		Doc:    "tz",
		Output: "build",
	}); err != nil {
		t.Fatalf("CreatePDF() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "build", "tz.pdf")); err != nil {
		t.Fatalf("expected tz.pdf: %v", err)
	}
}

func TestAIDraftWritesDraftFiles(t *testing.T) {
	root := t.TempDir()
	initProject(t, root, "single")
	if err := Generate(GenerateOptions{
		Doc:    "all",
		Output: root,
		SkipAI: true,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	projectDir := filepath.Join(root, "sample-project")
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
						"content": "= Сгенерированная секция\n\nTODO: Уточните детали проекта.\n",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Setenv("OPENROUTER_API_KEY", "test-key")
	t.Setenv("OPENROUTER_BASE_URL", server.URL)

	restore := chdir(t, root)
	defer restore()

	originalSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "01-overview.typ"))
	if err != nil {
		t.Fatal(err)
	}

	if err := AIDraft(AIDraftOptions{
		FromDoc:     "tz",
		Doc:         "pz",
		ProjectPath: projectDir,
		Model:       "test/model",
	}); err != nil {
		t.Fatalf("AIDraft() error = %v", err)
	}

	draft, err := os.ReadFile(filepath.Join(root, "docs", "pz", "drafts", "01-overview.typ"))
	if err != nil {
		t.Fatalf("expected generated draft: %v", err)
	}
	if !strings.Contains(string(draft), "Сгенерированная секция") {
		t.Fatalf("unexpected draft content: %s", string(draft))
	}

	currentSection, err := os.ReadFile(filepath.Join(root, "docs", "pz", "sections", "01-overview.typ"))
	if err != nil {
		t.Fatal(err)
	}
	if string(currentSection) != string(originalSection) {
		t.Fatalf("expected working section file to stay unchanged without --apply")
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

func initProject(t *testing.T, root, mode string) {
	t.Helper()
	if err := Init(InitOptions{
		Mode:   mode,
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
	cfg.Sources.TZPath = "./input/tz/main.typ"
	cfg.Sources.TeamTZPath = "./input/tz-team/main.typ"
	cfg.Sources.CodePaths = []string{}
	if err := config.Save(root, cfg); err != nil {
		t.Fatalf("Save() after init error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "input", "tz", "main.typ"), []byte("= ТЗ\n\nТекст технического задания.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "input", "tz-team", "main.typ"), []byte("= Командное ТЗ\n\nТекст командного ТЗ.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
