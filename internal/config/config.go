package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vasyukov1/hse-coursework-docs/internal/documents"
	"gopkg.in/yaml.v3"
)

const FileName = "term-paper.yaml"

type Config struct {
	Project      ProjectConfig
	Participants []Participant
	Supervisors  SupervisorsConfig
	Organization OrganizationConfig
	Sources      SourcesConfig
	Docs         map[string]DocumentConfig
	AI           AIConfig
}

type ProjectConfig struct {
	Name        string
	EnglishName string
	Code        string
	Summary     string
	Type        string
}

type Participant struct {
	Name  string
	Group string
}

type SupervisorsConfig struct {
	AgreedBy   Supervisor
	ApprovedBy Supervisor
}

type Supervisor struct {
	Name     string
	Position string
}

type OrganizationConfig struct {
	University string
	Faculty    string
	Program    string
	Year       int
}

type SourcesConfig struct {
	TZDir          string
	TeamTZDir      string
	CodeDir        string
	NotesFile      string
	SourcePriority string
}

type DocumentConfig struct {
	Enabled    bool
	Code       string
	Title      string
	AuthorMode string
	Annotation string
}

type AIConfig struct {
	Provider      string
	BaseURL       string
	DefaultModel  string
	APIKey        string
	Policy        []string
	StyleExamples map[string][]string
}

type userConfigFile struct {
	Project struct {
		Name        string `yaml:"name"`
		EnglishName string `yaml:"english_name"`
		Code        string `yaml:"code"`
		Summary     string `yaml:"summary"`
		Type        string `yaml:"type"`
	} `yaml:"project"`
	Student struct {
		Name  string `yaml:"name"`
		Group string `yaml:"group"`
	} `yaml:"student"`
	Supervisor struct {
		Name     string `yaml:"name"`
		Position string `yaml:"position"`
	} `yaml:"supervisor"`
	Approver struct {
		Name     string `yaml:"name"`
		Position string `yaml:"position"`
	} `yaml:"approver"`
	Organization struct {
		Name string `yaml:"name"`
	} `yaml:"organization"`
	Inputs struct {
		CodeDir        string `yaml:"code_dir"`
		TZDir          string `yaml:"tz_dir"`
		NotesFile      string `yaml:"notes_file"`
		SourcePriority string `yaml:"source_priority"`
	} `yaml:"inputs"`
	AI struct {
		Provider string `yaml:"provider"`
		BaseURL  string `yaml:"base_url"`
		Model    string `yaml:"model"`
		APIKey   string `yaml:"api_key"`
	} `yaml:"ai"`
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
			Code:        "05.01",
			Summary:     "Кратко опишите назначение проекта и что он делает.",
			Type:        "backend web-service",
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
			Year:       time.Now().Year(),
		},
		Sources: SourcesConfig{
			TZDir:          "./input/tz",
			TeamTZDir:      "./input/tz-team",
			CodeDir:        "./input/code",
			NotesFile:      "./input/notes.txt",
			SourcePriority: "balanced",
		},
		Docs: map[string]DocumentConfig{},
		AI: AIConfig{
			Provider:     "openrouter",
			BaseURL:      "https://openrouter.ai/api/v1",
			DefaultModel: "deepseek/deepseek-chat-v3-0324",
			APIKey:       "",
			Policy: []string{
				"Используй только факты из ТЗ и предоставленного проекта.",
				"Не копируй типовые формулировки из чужих работ; переформулируй под конкретный проект.",
				"Если фактов недостаточно, вставляй TODO вместо выдумывания.",
				"Если в документе нужны иллюстрации, пиши явные TODO в формате: TODO: вставить скрин ...",
				"Сохраняй структуру разделов, все заголовки шаблона и служебные Typst-импорты.",
				"Пиши развернуто: пользователь сможет удалить лишнее, но не должен дописывать пустые разделы с нуля.",
				"Не изменяй текст, который дан как базовый, например, названия используемой литературы",
				"Шаблон заполнения можешь найти по ссылке: https://github.com/vasyukov1/HSE-FCS-SE-2-year/blob/main/Term-Paper",
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
		cfg.Sources.TeamTZDir = "./input/tz-team"
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

func Load(projectRoot string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(projectRoot, FileName))
	if err != nil {
		return Config{}, err
	}

	var userCfg userConfigFile
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", FileName, err)
	}

	cfg := Default("single", documents.All())
	cfg.Project.Name = userCfg.Project.Name
	cfg.Project.EnglishName = userCfg.Project.EnglishName
	cfg.Project.Code = userCfg.Project.Code
	cfg.Project.Summary = userCfg.Project.Summary
	cfg.Project.Type = userCfg.Project.Type
	cfg.Participants = []Participant{{
		Name:  userCfg.Student.Name,
		Group: userCfg.Student.Group,
	}}
	cfg.Supervisors.AgreedBy.Name = userCfg.Supervisor.Name
	cfg.Supervisors.AgreedBy.Position = userCfg.Supervisor.Position
	cfg.Supervisors.ApprovedBy.Name = userCfg.Approver.Name
	cfg.Supervisors.ApprovedBy.Position = userCfg.Approver.Position
	if strings.TrimSpace(userCfg.Organization.Name) != "" {
		cfg.Organization.University = userCfg.Organization.Name
	}
	cfg.Sources.TZDir = userCfg.Inputs.TZDir
	cfg.Sources.TeamTZDir = "./input/tz-team"
	cfg.Sources.CodeDir = userCfg.Inputs.CodeDir
	cfg.Sources.NotesFile = userCfg.Inputs.NotesFile
	cfg.Sources.SourcePriority = userCfg.Inputs.SourcePriority
	cfg.AI.Provider = userCfg.AI.Provider
	cfg.AI.BaseURL = userCfg.AI.BaseURL
	cfg.AI.DefaultModel = userCfg.AI.Model
	cfg.AI.APIKey = userCfg.AI.APIKey
	return cfg, nil
}

func Save(projectRoot string, cfg Config) error {
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		return err
	}

	var userCfg userConfigFile
	userCfg.Project.Name = cfg.Project.Name
	userCfg.Project.EnglishName = cfg.Project.EnglishName
	userCfg.Project.Code = cfg.Project.Code
	userCfg.Project.Summary = cfg.Project.Summary
	userCfg.Project.Type = cfg.Project.Type
	if len(cfg.Participants) > 0 {
		userCfg.Student.Name = cfg.Participants[0].Name
		userCfg.Student.Group = cfg.Participants[0].Group
	}
	userCfg.Supervisor.Name = cfg.Supervisors.AgreedBy.Name
	userCfg.Supervisor.Position = cfg.Supervisors.AgreedBy.Position
	userCfg.Approver.Name = cfg.Supervisors.ApprovedBy.Name
	userCfg.Approver.Position = cfg.Supervisors.ApprovedBy.Position
	userCfg.Organization.Name = cfg.Organization.University
	userCfg.Inputs.CodeDir = cfg.Sources.CodeDir
	userCfg.Inputs.TZDir = cfg.Sources.TZDir
	userCfg.Inputs.NotesFile = cfg.Sources.NotesFile
	userCfg.Inputs.SourcePriority = cfg.Sources.SourcePriority
	userCfg.AI.Provider = cfg.AI.Provider
	userCfg.AI.BaseURL = cfg.AI.BaseURL
	userCfg.AI.Model = cfg.AI.DefaultModel
	userCfg.AI.APIKey = cfg.AI.APIKey

	data, err := yaml.Marshal(userCfg)
	if err != nil {
		return err
	}

	header := "# term-paper project configuration\n" +
		"# Положите основной ТЗ в директорию inputs.tz_dir\n" +
		"# Если нужен командный ПМИ, положите командное ТЗ в ./input/tz-team\n" +
		"# Положите архивы кода или папки репозиториев в директорию inputs.code_dir\n" +
		"# В inputs.notes_file можно добавить дополнительные требования, скриншоты и замечания\n"
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
	if len(cfg.Participants) == 0 {
		problems = append(problems, "student.name and student.group are required")
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
		problems = append(problems, "supervisor.name is required")
	}
	if strings.TrimSpace(cfg.Supervisors.AgreedBy.Position) == "" {
		problems = append(problems, "supervisor.position is required")
	}
	if strings.TrimSpace(cfg.Supervisors.ApprovedBy.Name) == "" {
		problems = append(problems, "approver.name is required")
	}
	if strings.TrimSpace(cfg.Supervisors.ApprovedBy.Position) == "" {
		problems = append(problems, "approver.position is required")
	}
	if strings.TrimSpace(cfg.Organization.University) == "" {
		problems = append(problems, "organization.name is required")
	}
	if strings.TrimSpace(cfg.Sources.TZDir) == "" {
		problems = append(problems, "inputs.tz_dir is required")
	}
	if strings.TrimSpace(cfg.Sources.CodeDir) == "" {
		problems = append(problems, "inputs.code_dir is required")
	}
	if strings.TrimSpace(cfg.Sources.NotesFile) == "" {
		problems = append(problems, "inputs.notes_file is required")
	}
	switch cfg.Sources.SourcePriority {
	case "tz", "code", "balanced":
	case "":
		problems = append(problems, "inputs.source_priority is required")
	default:
		problems = append(problems, "inputs.source_priority must be tz, code or balanced")
	}
	for _, spec := range documents.All() {
		docCfg, ok := cfg.Docs[spec.ID]
		if !ok {
			problems = append(problems, fmt.Sprintf("internal docs config for %s is missing", spec.ID))
			continue
		}
		if strings.TrimSpace(docCfg.Code) == "" {
			problems = append(problems, fmt.Sprintf("internal docs.%s.code is empty", spec.ID))
		}
		if strings.TrimSpace(docCfg.Title) == "" {
			problems = append(problems, fmt.Sprintf("internal docs.%s.title is empty", spec.ID))
		}
	}
	if strings.TrimSpace(cfg.AI.Provider) == "" {
		problems = append(problems, "ai.provider is required")
	}
	switch cfg.AI.Provider {
	case "openrouter", "openai", "anthropic", "google", "gemini":
	case "":
	default:
		problems = append(problems, "ai.provider must be openrouter, openai, anthropic, google or gemini")
	}
	if strings.TrimSpace(cfg.AI.BaseURL) == "" {
		problems = append(problems, "ai.base_url is required")
	}
	if strings.TrimSpace(cfg.AI.DefaultModel) == "" {
		problems = append(problems, "ai.model is required")
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
