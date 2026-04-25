package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/vasyukov1/term-paper/internal/ai"
	"github.com/vasyukov1/term-paper/internal/config"
	"github.com/vasyukov1/term-paper/internal/documents"
	"github.com/vasyukov1/term-paper/internal/templates"
	"github.com/vasyukov1/term-paper/internal/typst"
)

type GenerateOptions struct {
	Doc    string
	Output string
	SkipAI bool
}

type InitOptions struct {
	Mode   string
	Output string
}

type BundleTypstOptions struct {
	Input  string
	Output string
	Entry  string
}

type CreatePDFOptions struct {
	Doc    string
	Output string
	Watch  bool
}

type DoctorResult struct {
	Messages []string
}

type AIDraftOptions struct {
	FromDoc        string
	Doc            string
	Model          string
	ProjectPath    string
	ProjectArchive string
	Apply          bool
}

func Generate(opts GenerateOptions) error {
	projectRoot, cfg, err := loadProjectFromOutput(opts.Output)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}
	selected, err := selectDocsForGeneration(cfg, opts.Doc)
	if err != nil {
		return err
	}
	if err := templates.WriteProject(projectRoot, cfg, selected); err != nil {
		return err
	}

	if opts.SkipAI {
		fmt.Println("Typst project structure updated. AI generation skipped by --skip-ai.")
		return nil
	}

	return generateFromConfiguredSources(projectRoot, cfg, selected, "")
}

func Init(opts InitOptions) error {
	selected, err := documents.ResolveSelection("all", opts.Mode)
	if err != nil {
		return err
	}
	cfg := config.Default(opts.Mode, selected)
	if err := config.Validate(cfg); err != nil {
		return err
	}

	projectRoot, err := filepath.Abs(opts.Output)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "input", "code"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "input", "tz"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "input", "tz-team"), 0o755); err != nil {
		return err
	}
	if err := writeDefaultInputFile(filepath.Join(projectRoot, "input", "tz", "main.typ"), "= Техническое задание\n\nTODO: вставьте сюда основной текст ТЗ или замените файл своим Typst-документом.\n"); err != nil {
		return err
	}
	if err := writeDefaultInputFile(filepath.Join(projectRoot, "input", "tz-team", "main.typ"), "= Командное техническое задание\n\nTODO: если командного ТЗ нет, оставьте путь пустым в term-paper.yaml.\n"); err != nil {
		return err
	}
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}
	fmt.Printf("Created %s\n", filepath.Join(projectRoot, config.FileName))
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

func CreatePDF(opts CreatePDFOptions) error {
	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}
	docSpecs, err := resolveExistingDocs(cfg, opts.Doc)
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
		if opts.Watch {
			return typst.Watch(projectRoot, typst.MainFile(spec.ID), outputFile)
		}
		if err := typst.Compile(projectRoot, typst.MainFile(spec.ID), outputFile); err != nil {
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

	messages := []string{
		fmt.Sprintf("Project root: %s", projectRoot),
	}

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
	if err := ensureOptionalPath(resolveProjectPath(projectRoot, cfg.Sources.TZPath), "sources.tz_path"); err != nil {
		return DoctorResult{}, err
	}
	if strings.TrimSpace(cfg.Sources.TeamTZPath) != "" {
		if err := ensureOptionalPath(resolveProjectPath(projectRoot, cfg.Sources.TeamTZPath), "sources.team_tz_path"); err != nil {
			return DoctorResult{}, err
		}
	}
	for i, rawPath := range cfg.Sources.CodePaths {
		if strings.TrimSpace(rawPath) == "" {
			continue
		}
		if err := ensureOptionalPath(resolveProjectPath(projectRoot, rawPath), fmt.Sprintf("sources.code_paths[%d]", i)); err != nil {
			return DoctorResult{}, err
		}
	}

	for _, spec := range enabledSpecs(cfg) {
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

func AIDraft(opts AIDraftOptions) error {
	if opts.FromDoc != "tz" {
		return fmt.Errorf("only --from tz is supported in v1")
	}

	projectRoot, cfg, err := loadProject()
	if err != nil {
		return err
	}

	targetDocs, err := resolveExistingDocs(cfg, opts.Doc)
	if err != nil {
		return err
	}

	filtered := targetDocs[:0]
	for _, spec := range targetDocs {
		if spec.ID == opts.FromDoc {
			continue
		}
		filtered = append(filtered, spec)
	}
	if len(filtered) == 0 {
		return fmt.Errorf("no target documents selected after excluding %s", opts.FromDoc)
	}

	manualPaths := configuredCodePaths(projectRoot, cfg)
	if opts.ProjectPath != "" {
		manualPaths = []string{resolveProjectPath(projectRoot, opts.ProjectPath)}
	}
	if opts.ProjectArchive != "" {
		manualPaths = []string{resolveProjectPath(projectRoot, opts.ProjectArchive)}
	}

	return generateConfiguredDrafts(projectRoot, cfg, filtered, manualPaths, opts.Model, opts.Apply)
}

func generateDocDrafts(ctx context.Context, client ai.Client, projectRoot string, cfg config.Config, spec documents.Spec, model, sourceText, teamSourceText, projectContext string, apply bool) error {
	docCfg := cfg.Docs[spec.ID]
	for _, section := range spec.Sections {
		prompt := buildPrompt(cfg, spec, section, sourceText, teamSourceText, projectContext)
		content, err := client.Complete(ctx, model, []ai.Message{
			{
				Role: "system",
				Content: strings.Join(append([]string{
					"Ты помогаешь студенту писать документацию по курсовому проекту в формате Typst.",
					"Приоритет источников такой: основное ТЗ > командное ТЗ > код и структура проекта > заметки пользователя > стилевые примеры.",
					"Стилевые примеры можно использовать только как ориентир по тону и структуре. Нельзя переносить из них факты, названия сущностей и готовые абзацы.",
					"Пиши только на основе фактов из входных материалов.",
					"Не используй типовые обезличенные формулировки из чужих работ и не делай текст похожим на шаблон из интернета.",
					"Если информации недостаточно, оставляй локальный TODO и явно отмечай пробел.",
					"Если в документе нужны изображения, оставляй конкретные плейсхолдеры TODO на скриншоты.",
					"Верни только Typst-разметку для одного файла секции.",
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

		targetDir := filepath.Join(projectRoot, "docs", spec.Folder, "drafts")
		targetFile := filepath.Join(targetDir, section.FileName)
		if apply {
			targetDir = filepath.Join(projectRoot, "docs", spec.Folder, "sections")
			targetFile = filepath.Join(targetDir, section.FileName)
		}
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return err
		}
		normalized := ensureTrailingNewline(content)
		if err := os.WriteFile(targetFile, []byte(normalized), 0o644); err != nil {
			return err
		}
		fmt.Printf("Generated %s (%s)\n", targetFile, docCfg.Title)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func buildPrompt(cfg config.Config, spec documents.Spec, section documents.SectionSpec, sourceText, teamSourceText, projectContext string) string {
	var b strings.Builder
	styleExamples := ai.LoadReferenceExamples(cfg.AI.StyleExamples[spec.ID])

	b.WriteString(fmt.Sprintf("Документ: %s\n", spec.Title))
	b.WriteString(fmt.Sprintf("Секция: %s\n", section.Title))
	b.WriteString(fmt.Sprintf("Проект: %s\n", cfg.Project.Name))
	b.WriteString(fmt.Sprintf("Краткое описание: %s\n\n", cfg.Project.Summary))
	b.WriteString("Задача:\n")
	b.WriteString("Подготовь один законченный Typst-файл для указанной секции документации по курсовому проекту.\n\n")
	b.WriteString("Жёсткие правила:\n")
	b.WriteString("1. Основа текста — факты из основного ТЗ.\n")
	b.WriteString("2. Код и структура проекта нужны для проверки реализованности и конкретизации формулировок.\n")
	b.WriteString("3. Если ТЗ и код расходятся, не скрывай это: пиши нейтрально и ставь TODO там, где нужна ручная правка.\n")
	b.WriteString("4. Нельзя копировать ТЗ дословно большими фрагментами.\n")
	b.WriteString("5. Нельзя копировать стилевые примеры как готовый текст; примеры — только ориентир по тону и плотности.\n")
	b.WriteString("6. Нельзя выдумывать экраны, API, модули, алгоритмы, тесты, таблицы и метрики.\n")
	b.WriteString("7. Если сведений мало, пиши локальный TODO прямо в нужном месте, а не общую фразу в конце.\n")
	b.WriteString("8. Верни только Typst-разметку для одного файла секции.\n")
	b.WriteString("9. Начни с корректного заголовка секции.\n")
	b.WriteString("10. Если нужны изображения, вставляй конкретные TODO вида: TODO: вставить скрин ...\n")
	b.WriteString("11. Тон документа должен быть серьёзным, техническим и похожим на нормальную учебную документацию по ЕСПД.\n")
	b.WriteString(fmt.Sprintf("- Цель секции: %s\n\n", section.Prompt))
	if spec.ID == "pz" {
		b.WriteString("Это ПЗ. Сделай текст особенно конкретным и устойчивым к проверке на антиплагиат: меньше шаблонных формулировок, больше фактов из проекта и кода.\n\n")
	}
	b.WriteString("Рекомендуемый порядок работы:\n")
	b.WriteString("1. Извлеки факты из ТЗ.\n")
	b.WriteString("2. Сопоставь факты с кодом и структурой проекта.\n")
	b.WriteString("3. Сформируй связный текст раздела без воды.\n")
	b.WriteString("4. Оставь TODO только там, где действительно не хватает информации.\n\n")
	if len(cfg.Sources.Notes) > 0 {
		b.WriteString("Дополнительные указания пользователя:\n")
		for _, note := range cfg.Sources.Notes {
			b.WriteString("- " + note + "\n")
		}
		b.WriteString("\n")
	}
	if len(cfg.Sources.Screenshots) > 0 {
		b.WriteString("Заготовки для обязательных скриншотов:\n")
		for _, shot := range cfg.Sources.Screenshots {
			b.WriteString("- " + shot + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Исходное ТЗ:\n")
	b.WriteString(sourceText)
	if teamSourceText != "" {
		b.WriteString("\n\nКомандное ТЗ:\n")
		b.WriteString(teamSourceText)
	}
	if projectContext != "" {
		b.WriteString("\n\nКонтекст проекта:\n")
		b.WriteString(projectContext)
	}
	if styleExamples != "" {
		b.WriteString("\n\nСтилевые примеры из существующих документов пользователя:\n")
		b.WriteString("Используй их только как ориентир по типу формулировок, структуре и тону. Факты из них брать нельзя.\n")
		b.WriteString(styleExamples)
	}
	return b.String()
}

func generateConfiguredDrafts(projectRoot string, cfg config.Config, targetDocs []documents.Spec, codePaths []string, modelOverride string, apply bool) error {
	sourceText, err := loadPrimaryTZSource(projectRoot, cfg)
	if err != nil {
		return err
	}
	teamSourceText, err := loadTeamTZSource(projectRoot, cfg)
	if err != nil {
		return err
	}
	projectContext, err := ai.CollectProjectContexts(codePaths)
	if err != nil {
		return err
	}

	client := ai.Client{
		Provider: cfg.AI.Provider,
		BaseURL:  providerBaseURLFromEnv(cfg),
		APIKey:   firstNonEmpty(os.Getenv(cfg.AI.APIKeyEnv), cfg.AI.APIKey),
	}
	model := firstNonEmpty(modelOverride, cfg.AI.DefaultModel)
	if strings.TrimSpace(model) == "" {
		return fmt.Errorf("AI model is empty")
	}
	if strings.TrimSpace(client.APIKey) == "" {
		return fmt.Errorf("AI key is empty; set %s or ai.api_key in term-paper.yaml", cfg.AI.APIKeyEnv)
	}

	for _, spec := range targetDocs {
		if spec.ID == "tz" {
			continue
		}
		if err := generateDocDrafts(context.Background(), client, projectRoot, cfg, spec, model, sourceText, teamSourceTextForDoc(spec, teamSourceText), projectContext, apply); err != nil {
			return fmt.Errorf("generate %s draft: %w", spec.ID, err)
		}
	}
	return nil
}

func generateFromConfiguredSources(projectRoot string, cfg config.Config, selected []documents.Spec, modelOverride string) error {
	if strings.TrimSpace(cfg.Sources.TZPath) == "" {
		fmt.Println("Typst project structure updated. Fill sources.tz_path and run `term-paper generate` again for AI drafts.")
		return nil
	}
	apiKey := firstNonEmpty(os.Getenv(cfg.AI.APIKeyEnv), cfg.AI.APIKey)
	if strings.TrimSpace(apiKey) == "" {
		fmt.Println("Typst project structure updated. Fill ai.api_key or set the configured environment variable to enable AI drafts.")
		return nil
	}
	if err := generateConfiguredDrafts(projectRoot, cfg, selected, configuredCodePaths(projectRoot, cfg), modelOverride, false); err != nil {
		return err
	}
	fmt.Println("Typst project structure updated and AI drafts generated.")
	return nil
}

func enabledSpecs(cfg config.Config) []documents.Spec {
	var out []documents.Spec
	for _, spec := range documents.All() {
		if cfg.Docs[spec.ID].Enabled {
			out = append(out, spec)
		}
	}
	return out
}

func resolveExistingDocs(cfg config.Config, docArg string) ([]documents.Spec, error) {
	if docArg == "" || docArg == "all" {
		return enabledSpecs(cfg), nil
	}
	specs, err := documents.ResolveSelection(docArg, cfg.Project.Mode)
	if err != nil {
		return nil, err
	}
	filtered := specs[:0]
	for _, spec := range specs {
		docCfg, ok := cfg.Docs[spec.ID]
		if !ok {
			continue
		}
		if !docCfg.Enabled {
			return nil, fmt.Errorf("document %s is not enabled in this project", spec.ID)
		}
		filtered = append(filtered, spec)
	}
	return filtered, nil
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

func selectDocsForGeneration(cfg config.Config, docArg string) ([]documents.Spec, error) {
	if docArg == "" || docArg == "all" {
		return enabledSpecs(cfg), nil
	}
	specs, err := documents.ResolveSelection(docArg, cfg.Project.Mode)
	if err != nil {
		return nil, err
	}
	var selected []documents.Spec
	for _, spec := range specs {
		docCfg, ok := cfg.Docs[spec.ID]
		if !ok || !docCfg.Enabled {
			return nil, fmt.Errorf("document %s is not enabled in term-paper.yaml", spec.ID)
		}
		selected = append(selected, spec)
	}
	return selected, nil
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

func readSections(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".typ") {
			continue
		}
		names = append(names, entry.Name())
	}
	slices.Sort(names)

	var b strings.Builder
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		b.WriteString("\n\n# File: " + name + "\n")
		b.Write(data)
	}
	return strings.TrimSpace(b.String()), nil
}

func loadPrimaryTZSource(projectRoot string, cfg config.Config) (string, error) {
	if strings.TrimSpace(cfg.Sources.TZPath) != "" {
		path := resolveProjectPath(projectRoot, cfg.Sources.TZPath)
		text, err := ai.LoadSourceText(path)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(text) != "" {
			return text, nil
		}
	}
	return readSections(filepath.Join(projectRoot, "docs", "tz", "sections"))
}

func loadTeamTZSource(projectRoot string, cfg config.Config) (string, error) {
	if strings.TrimSpace(cfg.Sources.TeamTZPath) == "" {
		return "", nil
	}
	path := resolveProjectPath(projectRoot, cfg.Sources.TeamTZPath)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return ai.LoadSourceText(path)
}

func configuredCodePaths(projectRoot string, cfg config.Config) []string {
	paths := make([]string, 0, len(cfg.Sources.CodePaths))
	for _, raw := range cfg.Sources.CodePaths {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		paths = append(paths, resolveProjectPath(projectRoot, raw))
	}
	return paths
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
