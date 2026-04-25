package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vasyukov1/term-paper/internal/app"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "bundle-typst":
		return runBundleTypst(args[1:])
	case "generate":
		return runGenerate(args[1:])
	case "create-pdf":
		return runCreatePDF(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "ai":
		return runAI(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	mode := fs.String("mode", "single", "Project mode: single|team")
	output := fs.String("output", ".", "Destination directory for the generated config")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.Init(app.InitOptions{
		Mode:   *mode,
		Output: *output,
	})
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	doc := fs.String("doc", "all", "Document to generate: all|tz|pz|pmi|pmi-team|ro")
	output := fs.String("output", ".", "Project directory; should contain term-paper.yaml after init")
	skipAI := fs.Bool("skip-ai", false, "Only create the Typst project structure without AI drafts")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.Generate(app.GenerateOptions{
		Doc:    *doc,
		Output: *output,
		SkipAI: *skipAI,
	})
}

func runBundleTypst(args []string) error {
	fs := flag.NewFlagSet("bundle-typst", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	input := fs.String("input", "", "Path to a Typst file or directory with split Typst sources")
	output := fs.String("output", "", "Destination .typ file")
	entry := fs.String("entry", "main.typ", "Entry file inside --input when --input is a directory")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*input) == "" {
		return errors.New("--input is required")
	}
	if strings.TrimSpace(*output) == "" {
		return errors.New("--output is required")
	}

	return app.BundleTypst(app.BundleTypstOptions{
		Input:  *input,
		Output: *output,
		Entry:  *entry,
	})
}

func runCreatePDF(args []string) error {
	fs := flag.NewFlagSet("create-pdf", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	doc := fs.String("doc", "all", "Document to build: all|tz|pz|pmi|pmi-team|ro")
	output := fs.String("output", "build", "Directory for generated PDF files")
	watch := fs.Bool("watch", false, "Run typst watch for a single document")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.CreatePDF(app.CreatePDFOptions{
		Doc:    *doc,
		Output: *output,
		Watch:  *watch,
	})
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := app.Doctor()
	if err != nil {
		return err
	}

	for _, line := range result.Messages {
		fmt.Println(line)
	}
	return nil
}

func runAI(args []string) error {
	if len(args) == 0 {
		return errors.New("missing ai subcommand, expected: draft")
	}

	switch args[0] {
	case "draft":
		return runAIDraft(args[1:])
	default:
		return fmt.Errorf("unknown ai subcommand %q", args[0])
	}
}

func runAIDraft(args []string) error {
	fs := flag.NewFlagSet("ai draft", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fromDoc := fs.String("from", "tz", "Source document to ground generation from")
	doc := fs.String("doc", "all", "Target document: all|tz|pz|pmi|pmi-team|ro")
	projectPath := fs.String("project-path", "", "Optional path to the implementation project directory")
	projectArchive := fs.String("project-archive", "", "Optional path to a ZIP archive with the implementation project")
	model := fs.String("model", "", "Override model id for the OpenRouter-compatible API")
	apply := fs.Bool("apply", false, "Write generated drafts into working section files instead of drafts/")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *projectPath != "" && *projectArchive != "" {
		return errors.New("--project-path and --project-archive are mutually exclusive")
	}

	return app.AIDraft(app.AIDraftOptions{
		FromDoc:        *fromDoc,
		Doc:            *doc,
		Model:          *model,
		ProjectPath:    *projectPath,
		ProjectArchive: *projectArchive,
		Apply:          *apply,
	})
}

func printUsage() {
	fmt.Println(strings.TrimSpace(`
term-paper

Commands:
  term-paper init [--mode single|team] [--output <dir>]
  term-paper bundle-typst --input <file-or-dir> --output <file.typ> [--entry main.typ]
  term-paper generate [--doc all|tz|pz|pmi|pmi-team|ro] [--output <dir>] [--skip-ai]
  term-paper create-pdf [--doc all|tz|pz|pmi|pmi-team|ro] [--watch] [--output build]
  term-paper doctor
  term-paper ai draft --from tz [--doc ...] [--project-path <dir>|--project-archive <zip>] [--model <id>] [--apply]
`))
}
