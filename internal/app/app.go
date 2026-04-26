package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vasyukov1/hse-coursework-docs/internal/ai"
	"github.com/vasyukov1/hse-coursework-docs/internal/config"
	"github.com/vasyukov1/hse-coursework-docs/internal/documents"
	"github.com/vasyukov1/hse-coursework-docs/internal/templates"
	"github.com/vasyukov1/hse-coursework-docs/internal/typst"
)

type InitOptions struct {
	Output string
}

type BundleTypstOptions struct {
	Input  string
	Output string
	Entry  string
}

type GenerateDocOptions struct {
	Doc    string
	Output string
	Apply  bool
	Draft  bool
	SkipAI bool
}

type ImproveDocOptions struct {
	File   string
	Prompt string
	Apply  bool
}

type CreatePDFOptions struct {
	Doc    string
	Output string
	Watch  bool
}

type DoctorResult struct {
	Messages []string
}

func Init(opts InitOptions) error {
	selected, err := documents.ResolveSelection("all")
	if err != nil {
		return err
	}
	cfg := config.Default("single", selected)
	if err := config.Validate(cfg); err != nil {
		return err
	}

	projectRoot, err := filepath.Abs(opts.Output)
	if err != nil {
		return err
	}
	for _, path := range []string{
		filepath.Join(projectRoot, "input", "code"),
		filepath.Join(projectRoot, "input", "tz"),
		filepath.Join(projectRoot, "input", "tz-team"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}
	if err := writeDefaultInputFile(filepath.Join(projectRoot, "input", "tz", "main.typ"), "= Техническое задание\n\nTODO: вставьте сюда основной текст ТЗ или замените файл своими Typst-секциями.\n"); err != nil {
		return err
	}
	if err := writeDefaultInputFile(filepath.Join(projectRoot, "input", "tz-team", "main.typ"), "= Командное техническое задание\n\nTODO: используйте этот каталог только если нужен командный ПМИ.\n"); err != nil {
		return err
	}
	if err := writeDefaultInputFile(filepath.Join(projectRoot, "input", "notes.txt"), "TODO: добавьте важные детали проекта, расхождения между ТЗ и кодом, а также TODO для скриншотов.\n"); err != nil {
		return err
	}
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}
	fmt.Printf("Created %s\n", filepath.Join(projectRoot, config.FileName))
	fmt.Printf("Prepared input directories in %s\n", filepath.Join(projectRoot, "input"))
	return nil
}

func BundleTypst(opts BundleTypstOptions) error {
	content, err := typst.BundleSource(opts.Input, opts.Entry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opts.Output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(opts.Output, []byte(ensureTrailingNewline(content)), 0o644); err != nil {
		return err
	}
	fmt.Printf("Bundled Typst source written to %s\n", opts.Output)
	return nil
}

func GenerateDoc(opts GenerateDocOptions) error {
	projectRoot, cfg, err := loadProjectFromOutput(opts.Output)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}

	selected, err := selectSingleDoc(opts.Doc)
	if err != nil {
		return err
	}
	if err := templates.WriteProject(projectRoot, cfg, selected); err != nil {
		return err
	}
	if opts.SkipAI {
		fmt.Printf("Prepared Typst skeleton for %s without AI generation\n", selected[0].ID)
		return nil
	}

	return generateConfiguredDrafts(projectRoot, cfg, selected, cfg.AI.DefaultModel, !opts.Draft)
}

func ImproveDoc(opts ImproveDocOptions) error {
	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.File) == "" {
		return fmt.Errorf("--file is required")
	}
	if strings.TrimSpace(opts.Prompt) == "" {
		return fmt.Errorf("--prompt is required")
	}

	targetFile := resolveProjectPath(projectRoot, opts.File)
	current, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	primaryTZ, err := loadPrimaryTZSource(projectRoot, cfg)
	if err != nil {
		return err
	}
	codeContext, err := collectCodeContext(projectRoot, cfg)
	if err != nil {
		return err
	}
	notesText, err := loadNotes(projectRoot, cfg)
	if err != nil {
		return err
	}

	client := ai.Client{
		Provider: cfg.AI.Provider,
		BaseURL:  providerBaseURLFromEnv(cfg),
		APIKey:   configuredAPIKey(cfg),
	}
	if strings.TrimSpace(client.APIKey) == "" {
		return fmt.Errorf("AI key is empty; fill ai.api_key in term-paper.yaml or set a provider env var")
	}

	fmt.Printf("Improving %s...\n", targetFile)
	prompt := strings.Join([]string{
		"Улучши существующий Typst-файл курсовой документации НИУ ВШЭ.",
		"Верни только итоговый текст файла в формате Typst.",
		"Не добавляй пояснений, markdown-обёрток, вводных слов и комментариев вне текста документа.",
		"Сохрани все существующие #import, #include, заголовки и порядок разделов, если пользователь явно не попросил изменить структуру.",
		"Сохрани официальный стиль ЕСПД: плотный технический текст, проверяемые формулировки, без рекламного тона и разговорных оборотов.",
		"Заполняй раздел максимально подробно по ТЗ, коду и заметкам. Лучше оставить лишние редактируемые абзацы, чем короткий общий текст.",
		"Не выдумывай факты, версии, экраны, интеграции и метрики. Если данных не хватает, оставляй локальный TODO в нужном месте.",
		"Запрос пользователя: " + opts.Prompt,
		"\nТекущий текст файла:\n" + string(current),
		"\nОсновное ТЗ:\n" + primaryTZ,
		"\nКонтекст кода:\n" + projectContextOrStub(codeContext),
		"\nЗаметки пользователя:\n" + notesText,
	}, "\n\n")

	out, err := client.Complete(context.Background(), cfg.AI.DefaultModel, []ai.Message{
		{Role: "system", Content: "Верни только итоговый Typst-текст файла без пояснений. Структуру, заголовки и служебные импорты сохраняй как обязательный контракт."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return err
	}

	writePath := targetFile + ".improved.typ"
	if opts.Apply {
		writePath = targetFile
	}
	if err := os.WriteFile(writePath, []byte(ensureTrailingNewline(strings.TrimSpace(out))), 0o644); err != nil {
		return err
	}
	fmt.Printf("Improved file written to %s\n", writePath)
	return nil
}

func CreatePDF(opts CreatePDFOptions) error {
	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}
	docSpecs, err := resolveExistingDocs(projectRoot, cfg, opts.Doc)
	if err != nil {
		return err
	}
	if opts.Watch && len(docSpecs) != 1 {
		return fmt.Errorf("--watch requires a single document")
	}

	buildDir := filepath.Join(projectRoot, opts.Output)
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return err
	}

	for _, spec := range docSpecs {
		outputFile := filepath.Join(buildDir, spec.ID+".pdf")
		fmt.Printf("Building PDF for %s...\n", spec.ID)
		if opts.Watch {
			return typst.Watch(projectRoot, typst.MainFile(spec.Folder), outputFile)
		}
		if err := typst.Compile(projectRoot, typst.MainFile(spec.Folder), outputFile); err != nil {
			return fmt.Errorf("compile %s: %w", spec.ID, err)
		}
		fmt.Printf("Built %s\n", outputFile)
	}
	return nil
}

func Doctor() (DoctorResult, error) {
	projectRoot, cfg, err := loadProject()
	if err != nil {
		return DoctorResult{}, err
	}
	messages := []string{fmt.Sprintf("Project root: %s", projectRoot)}

	if err := typst.EnsureInstalled(); err != nil {
		return DoctorResult{}, err
	}
	messages = append(messages, "typst: ok")

	if err := config.Validate(cfg); err != nil {
		return DoctorResult{}, err
	}
	messages = append(messages, "config: ok")

	if err := ensurePath(filepath.Join(projectRoot, "shared", "typst", "core.typ")); err != nil {
		return DoctorResult{}, err
	}
	if err := ensureOptionalPath(resolveProjectPath(projectRoot, cfg.Sources.TZDir), "inputs.tz_dir"); err != nil {
		return DoctorResult{}, err
	}
	if err := ensureOptionalPath(resolveProjectPath(projectRoot, cfg.Sources.CodeDir), "inputs.code_dir"); err != nil {
		return DoctorResult{}, err
	}
	if err := ensureOptionalPath(resolveProjectPath(projectRoot, cfg.Sources.NotesFile), "inputs.notes_file"); err != nil {
		return DoctorResult{}, err
	}

	for _, spec := range existingDocSpecs(projectRoot) {
		docRoot := filepath.Join(projectRoot, "docs", spec.Folder)
		if err := ensurePath(filepath.Join(docRoot, "main.typ")); err != nil {
			return DoctorResult{}, err
		}
		for _, section := range spec.Sections {
			if err := ensurePath(filepath.Join(docRoot, "sections", section.FileName)); err != nil {
				return DoctorResult{}, fmt.Errorf("document %s is incomplete: %w", spec.ID, err)
			}
		}
		messages = append(messages, fmt.Sprintf("document %s: ok", spec.ID))
	}

	return DoctorResult{Messages: messages}, nil
}

func generateConfiguredDrafts(projectRoot string, cfg config.Config, targetDocs []documents.Spec, modelOverride string, apply bool) error {
	sourceText, err := loadPrimaryTZSource(projectRoot, cfg)
	if err != nil {
		return err
	}
	teamSourceText, err := loadTeamTZSource(projectRoot, cfg)
	if err != nil {
		return err
	}
	projectContext, err := collectCodeContext(projectRoot, cfg)
	if err != nil {
		return err
	}
	notesText, err := loadNotes(projectRoot, cfg)
	if err != nil {
		return err
	}

	client := ai.Client{
		Provider: cfg.AI.Provider,
		BaseURL:  providerBaseURLFromEnv(cfg),
		APIKey:   configuredAPIKey(cfg),
	}
	model := firstNonEmpty(modelOverride, cfg.AI.DefaultModel)
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("AI model is empty")
	}
	if strings.TrimSpace(client.APIKey) == "" {
		return fmt.Errorf("AI key is empty; fill ai.api_key in term-paper.yaml or set a provider env var")
	}

	for _, spec := range targetDocs {
		fmt.Printf("Preparing %s...\n", spec.ID)
		if err := generateDocDrafts(context.Background(), client, projectRoot, cfg, spec, model, sourceText, teamSourceTextForDoc(spec, teamSourceText), projectContext, notesText, apply); err != nil {
			return fmt.Errorf("generate %s draft: %w", spec.ID, err)
		}
	}
	return nil
}

func generateDocDrafts(ctx context.Context, client ai.Client, projectRoot string, cfg config.Config, spec documents.Spec, model, sourceText, teamSourceText, projectContext, notesText string, apply bool) error {
	for _, section := range spec.Sections {
		if !section.Generate || strings.TrimSpace(section.Prompt) == "" {
			fmt.Printf("Skipping %s/%s: static section\n", spec.ID, section.FileName)
			continue
		}

		fmt.Printf("Generating %s/%s...\n", spec.ID, section.FileName)
		targetFile := filepath.Join(projectRoot, "docs", spec.Folder, "sections", section.FileName)
		sectionTemplate, err := os.ReadFile(targetFile)
		if err != nil {
			return err
		}
		prompt := buildPrompt(cfg, spec, section, string(sectionTemplate), sourceText, teamSourceText, projectContext, notesText)
		content, err := client.Complete(ctx, model, []ai.Message{
			{
				Role: "system",
				Content: strings.Join(append([]string{
					"Ты пишешь курсовую программную документацию НИУ ВШЭ в формате Typst по ЕСПД.",
					"Главный источник знаний определяется полем inputs.source_priority: tz, code или balanced.",
					"Текущий шаблон секции является обязательным контрактом: сохрани #import, верхний заголовок, все подзаголовки и их порядок.",
					"Можно заменять TODO на содержательный текст и добавлять дополнительные абзацы, списки, таблицы и подпункты ниже существующих заголовков.",
					"Нельзя удалять обязательные пункты шаблона, переименовывать разделы, менять стиль документа или добавлять markdown-обёртки.",
					"Используй стилевые примеры только как ориентир по плотности, тону и структуре. Нельзя переносить из них факты, названия сущностей и готовые абзацы.",
					"Верни только итоговый Typst-текст секции без пояснений модели.",
					"Не выдумывай факты. Если данных не хватает, оставляй локальный TODO внутри текста секции.",
					"Генерируй подробный материал: проще удалить лишние корректные абзацы, чем дописывать пустой раздел с нуля.",
					"Стиль должен быть официальным, техническим, проверяемым и уникальным, без плагиата и разговорных формулировок.",
				}, cfg.AI.Policy...), "\n"),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		})
		if err != nil {
			return err
		}

		targetDir := filepath.Join(projectRoot, "docs", spec.Folder, "sections")
		if !apply {
			targetDir = filepath.Join(projectRoot, "docs", spec.Folder, "drafts")
			targetFile = filepath.Join(targetDir, section.FileName)
		}
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(targetFile, []byte(ensureTrailingNewline(strings.TrimSpace(content))), 0o644); err != nil {
			return err
		}
		fmt.Printf("Generated %s\n", targetFile)
	}
	return nil
}

func buildPrompt(cfg config.Config, spec documents.Spec, section documents.SectionSpec, sectionTemplate, sourceText, teamSourceText, projectContext, notesText string) string {
	var b strings.Builder
	styleExamples := ai.LoadReferenceExamples(cfg.AI.StyleExamples[spec.ID])

	b.WriteString("Задача: подготовь итоговый текст одной секции документа в формате Typst.\n\n")
	b.WriteString("Документ: " + spec.Title + "\n")
	b.WriteString("Секция: " + section.Title + "\n")
	b.WriteString("Проект: " + cfg.Project.Name + "\n")
	b.WriteString("Название на английском: " + cfg.Project.EnglishName + "\n")
	b.WriteString("Код проекта: " + cfg.Project.Code + "\n")
	b.WriteString("Тип проекта: " + cfg.Project.Type + "\n")
	b.WriteString("Краткое описание: " + cfg.Project.Summary + "\n")
	b.WriteString("Студент: " + cfg.Participants[0].Name + ", " + cfg.Participants[0].Group + "\n\n")

	b.WriteString("Обязательный шаблон секции:\n")
	b.WriteString(sectionTemplate + "\n\n")

	b.WriteString("Требования к результату:\n")
	b.WriteString("1. Верни только итоговый текст секции в формате Typst.\n")
	b.WriteString("2. Сохрани все #import, верхний заголовок, подзаголовки и порядок пунктов из обязательного шаблона секции.\n")
	b.WriteString("3. Заполняй TODO и пустые места содержательным текстом, но не удаляй обязательные пункты шаблона.\n")
	b.WriteString("4. Не добавляй вводных слов, не пиши, что ты что-то сгенерировал, не используй markdown.\n")
	b.WriteString("5. Не выдумывай факты, технологии, интерфейсы, показатели, тесты и экраны, которых нет в ТЗ, коде или заметках.\n")
	b.WriteString("6. Если для полноценного описания нужны скриншоты, вставляй только TODO вида: TODO: вставить скрин ...\n")
	b.WriteString("7. Раздел литературы и служебные списки ГОСТ оставляй без изменений, если генерируемая секция не требует проектной адаптации.\n")
	b.WriteString("8. Пиши официально и технически, но без перегруженных сложных фраз.\n")
	b.WriteString("9. Делай текст подробным: раскрывай назначения, ограничения, входные и выходные данные, критерии проверки, ошибки и пользовательские сценарии, когда это применимо.\n")
	b.WriteString("10. Для крупных функциональных разделов добавляй проектно-специфичные подпункты третьего уровня, если это не ломает обязательную структуру.\n\n")

	b.WriteString("Что нужно раскрыть в этой секции:\n" + section.Prompt + "\n\n")
	b.WriteString("Приоритет источников: " + cfg.Sources.SourcePriority + "\n\n")

	b.WriteString("Основное ТЗ:\n" + sourceText + "\n")
	if teamSourceText != "" {
		b.WriteString("\nКомандное ТЗ:\n" + teamSourceText + "\n")
	}
	if strings.TrimSpace(projectContext) != "" {
		b.WriteString("\nКонтекст кода:\n" + projectContext + "\n")
	}
	if strings.TrimSpace(notesText) != "" {
		b.WriteString("\nЗаметки пользователя:\n" + notesText + "\n")
	}
	if styleExamples != "" {
		b.WriteString("\nСтилевые примеры:\n")
		b.WriteString("Используй их только как ориентир по плотности и тону текста. Переносить факты и готовые фразы нельзя.\n")
		b.WriteString(styleExamples + "\n")
	}

	return b.String()
}

func existingDocSpecs(projectRoot string) []documents.Spec {
	var out []documents.Spec
	for _, spec := range documents.All() {
		if _, err := os.Stat(filepath.Join(projectRoot, "docs", spec.Folder, "main.typ")); err == nil {
			out = append(out, spec)
		}
	}
	return out
}

func resolveExistingDocs(projectRoot string, cfg config.Config, docArg string) ([]documents.Spec, error) {
	if docArg == "" || docArg == "all" {
		specs := existingDocSpecs(projectRoot)
		if len(specs) == 0 {
			return nil, fmt.Errorf("no generated documents found; run `term-paper generate-doc --doc <...>` first")
		}
		return specs, nil
	}

	specs, err := documents.ResolveSelection(docArg)
	if err != nil {
		return nil, err
	}
	for _, spec := range specs {
		if _, err := os.Stat(filepath.Join(projectRoot, "docs", spec.Folder, "main.typ")); err != nil {
			return nil, fmt.Errorf("document %s was not created yet; run `term-paper generate-doc --doc %s` first", spec.ID, spec.ID)
		}
	}
	return specs, nil
}

func selectSingleDoc(docArg string) ([]documents.Spec, error) {
	if strings.TrimSpace(docArg) == "" || docArg == "all" {
		return nil, fmt.Errorf("use one document id: pz, pmi, ro or pmi-team")
	}
	specs, err := documents.ResolveSelection(docArg)
	if err != nil {
		return nil, err
	}
	if len(specs) != 1 {
		return nil, fmt.Errorf("exactly one document is required")
	}
	if specs[0].ID == "tz" {
		return nil, fmt.Errorf("direct generation for tz is not supported")
	}
	return specs, nil
}

func loadProject() (string, config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", config.Config{}, err
	}
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return "", config.Config{}, err
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return "", config.Config{}, err
	}
	return projectRoot, cfg, nil
}

func loadProjectFromOutput(output string) (string, config.Config, error) {
	projectRoot, err := filepath.Abs(output)
	if err != nil {
		return "", config.Config{}, err
	}
	cfg, err := config.Load(projectRoot)
	if err == nil {
		return projectRoot, cfg, nil
	}
	if !os.IsNotExist(err) {
		return "", config.Config{}, err
	}
	return "", config.Config{}, fmt.Errorf("could not find %s in %s; run `term-paper init --output %s` first", config.FileName, projectRoot, projectRoot)
}

func ensurePath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("missing %s", path)
	}
	return nil
}

func ensureOptionalPath(path, field string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s points to a missing path: %s", field, path)
	}
	return nil
}

func loadPrimaryTZSource(projectRoot string, cfg config.Config) (string, error) {
	return typst.BundleSource(resolveProjectPath(projectRoot, cfg.Sources.TZDir), "main.typ")
}

func loadTeamTZSource(projectRoot string, cfg config.Config) (string, error) {
	path := resolveProjectPath(projectRoot, cfg.Sources.TeamTZDir)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return typst.BundleSource(path, "main.typ")
}

func collectCodeContext(projectRoot string, cfg config.Config) (string, error) {
	codeDir := resolveProjectPath(projectRoot, cfg.Sources.CodeDir)
	entries, err := os.ReadDir(codeDir)
	if err != nil {
		return "", err
	}
	var paths []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		paths = append(paths, filepath.Join(codeDir, entry.Name()))
	}
	sort.Strings(paths)
	return ai.CollectProjectContexts(paths)
}

func loadNotes(projectRoot string, cfg config.Config) (string, error) {
	path := resolveProjectPath(projectRoot, cfg.Sources.NotesFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func resolveProjectPath(projectRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}

func teamSourceTextForDoc(spec documents.Spec, teamSourceText string) string {
	if spec.ID == "pmi-team" {
		return teamSourceText
	}
	return ""
}

func configuredAPIKey(cfg config.Config) string {
	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "openai":
		return firstNonEmpty(os.Getenv("OPENAI_API_KEY"), cfg.AI.APIKey)
	case "anthropic":
		return firstNonEmpty(os.Getenv("ANTHROPIC_API_KEY"), cfg.AI.APIKey)
	default:
		return firstNonEmpty(os.Getenv("OPENROUTER_API_KEY"), cfg.AI.APIKey)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func providerBaseURL(cfg config.Config) string {
	if strings.TrimSpace(cfg.AI.BaseURL) != "" {
		return cfg.AI.BaseURL
	}
	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "openai":
		return "https://api.openai.com/v1"
	case "anthropic":
		return "https://api.anthropic.com/v1"
	default:
		return "https://openrouter.ai/api/v1"
	}
}

func providerBaseURLFromEnv(cfg config.Config) string {
	switch strings.ToLower(strings.TrimSpace(cfg.AI.Provider)) {
	case "openai":
		return firstNonEmpty(os.Getenv("OPENAI_BASE_URL"), providerBaseURL(cfg))
	case "anthropic":
		return firstNonEmpty(os.Getenv("ANTHROPIC_BASE_URL"), providerBaseURL(cfg))
	default:
		return firstNonEmpty(os.Getenv("OPENROUTER_BASE_URL"), providerBaseURL(cfg))
	}
}

func ensureTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func writeDefaultInputFile(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func projectContextOrStub(text string) string {
	if strings.TrimSpace(text) == "" {
		return "Контекст кода отсутствует."
	}
	return text
}
