package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vasyukov1/hse-coursework-docs/internal/app"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp("")
		return nil
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "generate-doc":
		return runGenerateDoc(args[1:])
	case "create-pdf":
		return runCreatePDF(args[1:])
	case "bundle-typst":
		return runBundleTypst(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "improve-doc":
		return runImproveDoc(args[1:])
	case "help", "-h", "--help":
		topic := ""
		if len(args) > 1 {
			topic = args[1]
		}
		printHelp(topic)
		return nil
	case "generate":
		return errors.New("command `generate` was removed; use `term-paper generate-doc --doc <pz|pmi|ro|pmi-team>`")
	default:
		return fmt.Errorf("unknown command %q; use `term-paper help`", args[0])
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	output := fs.String("output", ".", "Directory where term-paper.yaml and input/ will be created")
	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.Init(app.InitOptions{
		Output: *output,
	})
}

func runGenerateDoc(args []string) error {
	fs := flag.NewFlagSet("generate-doc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	doc := fs.String("doc", "", "Document to generate: pz|pmi|ro|pmi-team")
	output := fs.String("output", ".", "Project directory with term-paper.yaml")
	apply := fs.Bool("apply", false, "Write generated text directly into docs/<doc>/sections instead of docs/<doc>/drafts")
	skipAI := fs.Bool("skip-ai", false, "Only create the Typst document skeleton without AI generation")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*doc) == "" {
		return errors.New("--doc is required")
	}

	return app.GenerateDoc(app.GenerateDocOptions{
		Doc:    *doc,
		Output: *output,
		Apply:  *apply,
		SkipAI: *skipAI,
	})
}

func runCreatePDF(args []string) error {
	fs := flag.NewFlagSet("create-pdf", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	doc := fs.String("doc", "all", "Document to build: all|tz|pz|pmi|ro|pmi-team")
	output := fs.String("output", "build", "Directory for PDF files")
	watch := fs.Bool("watch", false, "Run typst watch for one selected document")
	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.CreatePDF(app.CreatePDFOptions{
		Doc:    *doc,
		Output: *output,
		Watch:  *watch,
	})
}

func runBundleTypst(args []string) error {
	fs := flag.NewFlagSet("bundle-typst", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	input := fs.String("input", "", "Path to a Typst file or a directory with split Typst sources")
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

func runImproveDoc(args []string) error {
	fs := flag.NewFlagSet("improve-doc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	file := fs.String("file", "", "Path to the .typ file that should be improved")
	prompt := fs.String("prompt", "", "What exactly should be improved")
	apply := fs.Bool("apply", false, "Overwrite the target file instead of writing <file>.improved.typ")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*file) == "" {
		return errors.New("--file is required")
	}
	if strings.TrimSpace(*prompt) == "" {
		return errors.New("--prompt is required")
	}

	return app.ImproveDoc(app.ImproveDocOptions{
		File:   *file,
		Prompt: *prompt,
		Apply:  *apply,
	})
}

func printHelp(topic string) {
	switch strings.TrimSpace(topic) {
	case "", "overview":
		fmt.Println(strings.TrimSpace(`
term-paper

Инструмент для подготовки курсовой документации в Typst.

Основной сценарий:
  1. term-paper init
  2. заполнить term-paper.yaml
  3. положить ТЗ в input/tz, код в input/code, заметки в input/notes.txt
  4. term-paper generate-doc --doc pz
  5. проверить drafts/ или sections/
  6. term-paper create-pdf --doc pz

Команды:
  term-paper init [--output <dir>]
  term-paper generate-doc --doc <pz|pmi|ro|pmi-team> [--output <dir>] [--apply] [--skip-ai]
  term-paper create-pdf [--doc all|tz|pz|pmi|ro|pmi-team] [--output build] [--watch]
  term-paper bundle-typst --input <file-or-dir> --output <file.typ> [--entry main.typ]
  term-paper improve-doc --file <path.typ> --prompt "<что улучшить>" [--apply]
  term-paper doctor
  term-paper help [command]

Подсказки:
  term-paper help init
  term-paper help generate-doc
  term-paper help create-pdf
  term-paper help bundle-typst
  term-paper help improve-doc
`))
	case "init":
		fmt.Println(strings.TrimSpace(`
term-paper init

Что делает:
  - создаёт term-paper.yaml
  - создаёт input/code
  - создаёт input/tz
  - создаёт input/tz-team
  - создаёт input/notes.txt

Пример:
  term-paper init
  term-paper init --output ./my-coursework
`))
	case "generate-doc":
		fmt.Println(strings.TrimSpace(`
term-paper generate-doc

Что делает:
  - создаёт Typst-шаблон выбранного документа
  - читает ТЗ, код и заметки
  - генерирует только один документ за запуск
  - сохраняет результат в docs/<doc>/drafts
  - с --apply пишет сразу в docs/<doc>/sections

Документы:
  pz        пояснительная записка
  pmi       программа и методика испытаний
  ro        руководство оператора
  pmi-team  командное ПМИ, использует дополнительно input/tz-team

Примеры:
  term-paper generate-doc --doc pz
  term-paper generate-doc --doc pmi --apply
  term-paper generate-doc --doc ro --skip-ai
`))
	case "create-pdf":
		fmt.Println(strings.TrimSpace(`
term-paper create-pdf

Что делает:
  - собирает PDF через локальный typst
  - может собрать один документ или все уже созданные документы

Примеры:
  term-paper create-pdf
  term-paper create-pdf --doc pz
  term-paper create-pdf --doc pmi --watch
`))
	case "bundle-typst":
		fmt.Println(strings.TrimSpace(`
term-paper bundle-typst

Что делает:
  - собирает Typst из нескольких секций в один .typ файл
  - удобно для повторной генерации или отправки ТЗ как одного файла

Примеры:
  term-paper bundle-typst --input ./input/tz --output ./input/tz-bundled.typ
  term-paper bundle-typst --input ./some-doc --entry body.typ --output ./one-file.typ
`))
	case "improve-doc":
		fmt.Println(strings.TrimSpace(`
term-paper improve-doc

Что делает:
  - берёт существующий Typst-файл
  - улучшает его по вашему текстовому запросу
  - по умолчанию пишет результат в <file>.improved.typ

Примеры:
  term-paper improve-doc --file docs/pz/sections/03-tech.typ --prompt "сделай раздел подробнее и формальнее"
  term-paper improve-doc --file docs/ro/sections/03-run.typ --prompt "добавь больше шагов установки" --apply
`))
	case "doctor":
		fmt.Println(strings.TrimSpace(`
term-paper doctor

Что проверяет:
  - установлен ли typst
  - найден ли term-paper.yaml
  - валиден ли конфиг
  - существуют ли input-пути
  - не повреждены ли уже созданные Typst-документы

Пример:
  term-paper doctor
`))
	default:
		fmt.Printf("Unknown help topic %q\n", topic)
	}
}
