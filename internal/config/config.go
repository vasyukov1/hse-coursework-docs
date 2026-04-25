package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vasyukov1/term-paper/internal/documents"
	"gopkg.in/yaml.v3"
)

const FileName = "term-paper.yaml"

type Config struct {
	Project      ProjectConfig             `yaml:"project"`
	Participants []Participant             `yaml:"participants"`
	Supervisors  SupervisorsConfig         `yaml:"supervisors"`
	Organization OrganizationConfig        `yaml:"organization"`
	Sources      SourcesConfig             `yaml:"sources"`
	Docs         map[string]DocumentConfig `yaml:"docs"`
	AI           AIConfig                  `yaml:"ai"`
}

type ProjectConfig struct {
	Name        string `yaml:"name"`
	EnglishName string `yaml:"english_name"`
	Code        string `yaml:"code"`
	Summary     string `yaml:"summary"`
	Type        string `yaml:"type"`
	Mode        string `yaml:"mode"`
}

type Participant struct {
	Name  string `yaml:"name"`
	Group string `yaml:"group"`
}

type SupervisorsConfig struct {
	AgreedBy   Supervisor `yaml:"agreed_by"`
	ApprovedBy Supervisor `yaml:"approved_by"`
}

type Supervisor struct {
	Name     string `yaml:"name"`
	Position string `yaml:"position"`
}

type OrganizationConfig struct {
	University string `yaml:"university"`
	Faculty    string `yaml:"faculty"`
	Program    string `yaml:"program"`
	Year       int    `yaml:"year"`
}

type DocumentConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Code       string `yaml:"code"`
	Title      string `yaml:"title"`
	AuthorMode string `yaml:"author_mode"`
	Annotation string `yaml:"annotation"`
}

type SourcesConfig struct {
	TZPath      string   `yaml:"tz_path"`
	TeamTZPath  string   `yaml:"team_tz_path"`
	CodePaths   []string `yaml:"code_paths"`
	Notes       []string `yaml:"notes"`
	Screenshots []string `yaml:"screenshots"`
}

type AIConfig struct {
	Provider      string              `yaml:"provider"`
	BaseURL       string              `yaml:"base_url"`
	DefaultModel  string              `yaml:"default_model"`
	APIKey        string              `yaml:"api_key"`
	APIKeyEnv     string              `yaml:"api_key_env"`
	Policy        []string            `yaml:"policy"`
	StyleExamples map[string][]string `yaml:"style_examples"`
}

func Default(mode string, selected []documents.Spec) Config {
	selectedSet := map[string]bool{}
	for _, spec := range selected {
		selectedSet[spec.ID] = true
	}

	cfg := Config{
		Project: ProjectConfig{
			Name:        "Название проекта",
			EnglishName: "Project title in English",
			Code:        "05.01-01",
			Summary:     "Кратко опишите назначение, пользователей и ожидаемый результат проекта.",
			Type:        "web-service",
			Mode:        mode,
		},
		Participants: []Participant{
			{Name: "И. О. Фамилия", Group: "БПИ000"},
		},
		Supervisors: SupervisorsConfig{
			AgreedBy: Supervisor{
				Name:     "Ф. И. О. Научного руководителя",
				Position: "Научный руководитель, должность",
			},
			ApprovedBy: Supervisor{
				Name:     "Н. А. Павлочев",
				Position: "Академический руководитель образовательной программы \"Программная инженерия\", старший преподаватель департамента программной инженерии",
			},
		},
		Organization: OrganizationConfig{
			University: "Национальный исследовательский университет \"Высшая школа экономики\"",
			Faculty:    "Факультет компьютерных наук",
			Program:    "Образовательная программа \"Программная инженерия\"",
			Year:       2026,
		},
		Sources: SourcesConfig{
			TZPath:    "./input/tz/main.typ",
			CodePaths: []string{},
			Notes: []string{
				"Проект: сервис для проверки курсовых работ.",
				"ПЗ должно быть максимально привязано к ТЗ и к реально реализованному коду.",
			},
			Screenshots: []string{
				"TODO: вставить скрин экрана авторизации",
				"TODO: вставить скрин главного экрана",
			},
		},
		Docs: map[string]DocumentConfig{},
		AI: AIConfig{
			Provider:     "openrouter",
			BaseURL:      "https://openrouter.ai/api/v1",
			DefaultModel: "deepseek/deepseek-chat-v3-0324",
			APIKey:       "",
			APIKeyEnv:    "OPENROUTER_API_KEY",
			Policy: []string{
				"Используй только факты из ТЗ и предоставленного проекта.",
				"Не копируй типовые формулировки из чужих работ; переформулируй под конкретный проект.",
				"Если фактов недостаточно, вставляй TODO вместо выдумывания.",
				"Если в документе нужны иллюстрации, пиши явные TODO в формате: TODO: вставить скрин ...",
				"Сохраняй структуру разделов и пиши в формате Typst.",
			},
			StyleExamples: map[string][]string{
				"tz": {
					"https://raw.githubusercontent.com/vasyukov1/HSE-FCS-SE-2-year/main/Term-Paper/%D0%A2%D0%97/body.typ",
				},
				"pz": {
					"https://raw.githubusercontent.com/vasyukov1/HSE-FCS-SE-2-year/main/Term-Paper/%D0%9F%D0%97/body.typ",
				},
				"pmi": {
					"https://raw.githubusercontent.com/vasyukov1/HSE-FCS-SE-2-year/main/Term-Paper/%D0%9F%D0%9C%D0%98/body.typ",
				},
				"pmi-team": {
					"https://raw.githubusercontent.com/vasyukov1/HSE-FCS-SE-2-year/main/Term-Paper/%D0%9F%D0%9C%D0%98%20%D0%BA%D0%BE%D0%BC%D0%B0%D0%BD%D0%B4%D0%BD%D0%BE%D0%B5/body.typ",
				},
				"ro": {
					"https://raw.githubusercontent.com/vasyukov1/HSE-FCS-SE-2-year/main/Term-Paper/%D0%A0%D0%9E/body.typ",
				},
			},
		},
	}

	if mode == "team" {
		cfg.Participants = []Participant{
			{Name: "И. О. Фамилия", Group: "БПИ000"},
			{Name: "И. О. Второй участник", Group: "БПИ000"},
		}
		cfg.Sources.TeamTZPath = "./input/tz-team/main.typ"
	}

	for _, spec := range documents.All() {
		cfg.Docs[spec.ID] = DocumentConfig{
			Enabled:    selectedSet[spec.ID],
			Code:       defaultDocCode(spec.ID),
			Title:      spec.Title,
			AuthorMode: spec.DefaultAuthorMode,
			Annotation: defaultAnnotation(spec.ID),
		}
	}

	return cfg
}

func defaultDocCode(docID string) string {
	switch docID {
	case "tz":
		return "RU.17701729.05.01-01 ТЗ 01-1"
	case "pz":
		return "RU.17701729.05.01-01 81 01-1"
	case "pmi", "pmi-team":
		return "RU.17701729.05.01-01 51 01-1"
	case "ro":
		return "RU.17701729.05.01-01 34 01-1"
	default:
		return "RU.17701729.05.01-01"
	}
}

func defaultAnnotation(docID string) string {
	switch docID {
	case "tz":
		return "Техническое задание фиксирует исходные требования к программному продукту, состав работ и порядок приемки."
	case "pz":
		return "Пояснительная записка описывает назначение проекта, принятые технические решения и ожидаемые результаты."
	case "pmi", "pmi-team":
		return "Программа и методика испытаний определяет цели, средства и порядок проверки работоспособности программного продукта."
	case "ro":
		return ""
	default:
		return ""
	}
}

func Load(projectRoot string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, FileName))
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", FileName, err)
	}

	return cfg, nil
}

func Save(projectRoot string, cfg Config) error {
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	header := "# term-paper project configuration\n"
	return os.WriteFile(filepath.Join(projectRoot, FileName), append([]byte(header), data...), 0o644)
}

func Validate(cfg Config) error {
	var problems []string
	if strings.TrimSpace(cfg.Project.Name) == "" {
		problems = append(problems, "project.name is required")
	}
	if strings.TrimSpace(cfg.Project.EnglishName) == "" {
		problems = append(problems, "project.english_name is required")
	}
	if strings.TrimSpace(cfg.Project.Code) == "" {
		problems = append(problems, "project.code is required")
	}
	if cfg.Project.Mode != "single" && cfg.Project.Mode != "team" {
		problems = append(problems, "project.mode must be single or team")
	}
	if len(cfg.Participants) == 0 {
		problems = append(problems, "participants must contain at least one student")
	}
	for i, participant := range cfg.Participants {
		if strings.TrimSpace(participant.Name) == "" {
			problems = append(problems, fmt.Sprintf("participants[%d].name is required", i))
		}
		if strings.TrimSpace(participant.Group) == "" {
			problems = append(problems, fmt.Sprintf("participants[%d].group is required", i))
		}
	}
	if strings.TrimSpace(cfg.Supervisors.AgreedBy.Name) == "" {
		problems = append(problems, "supervisors.agreed_by.name is required")
	}
	if strings.TrimSpace(cfg.Supervisors.ApprovedBy.Name) == "" {
		problems = append(problems, "supervisors.approved_by.name is required")
	}
	if strings.TrimSpace(cfg.Organization.University) == "" {
		problems = append(problems, "organization.university is required")
	}
	if cfg.Organization.Year == 0 {
		problems = append(problems, "organization.year is required")
	}

	for _, spec := range documents.All() {
		docCfg, ok := cfg.Docs[spec.ID]
		if !ok {
			problems = append(problems, fmt.Sprintf("docs.%s entry is missing", spec.ID))
			continue
		}
		if strings.TrimSpace(docCfg.Code) == "" {
			problems = append(problems, fmt.Sprintf("docs.%s.code is required", spec.ID))
		}
		if strings.TrimSpace(docCfg.Title) == "" {
			problems = append(problems, fmt.Sprintf("docs.%s.title is required", spec.ID))
		}
		if docCfg.AuthorMode != "single" && docCfg.AuthorMode != "team" {
			problems = append(problems, fmt.Sprintf("docs.%s.author_mode must be single or team", spec.ID))
		}
	}
	if strings.TrimSpace(cfg.Sources.TZPath) == "" {
		problems = append(problems, "sources.tz_path is required")
	}
	if strings.TrimSpace(cfg.AI.Provider) == "" {
		problems = append(problems, "ai.provider is required")
	}
	switch cfg.AI.Provider {
	case "openrouter", "openai", "anthropic":
	case "":
	default:
		problems = append(problems, "ai.provider must be openrouter, openai or anthropic")
	}
	if strings.TrimSpace(cfg.AI.BaseURL) == "" {
		problems = append(problems, "ai.base_url is required")
	}
	if strings.TrimSpace(cfg.AI.DefaultModel) == "" {
		problems = append(problems, "ai.default_model is required")
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func EnabledDocIDs(cfg Config) []string {
	ids := make([]string, 0, len(cfg.Docs))
	for _, spec := range documents.All() {
		if cfg.Docs[spec.ID].Enabled {
			ids = append(ids, spec.ID)
		}
	}
	return ids
}

func FindProjectRoot(start string) (string, error) {
	current := start
	for {
		if _, err := os.Stat(filepath.Join(current, FileName)); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("could not find %s in %s or its parents", FileName, start)
		}
		current = parent
	}
}
