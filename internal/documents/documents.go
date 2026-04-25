package documents

import (
	"fmt"
	"slices"
	"strings"
)

type SectionSpec struct {
	FileName string
	Title    string
	Prompt   string
}

type Spec struct {
	ID                string
	Folder            string
	Title             string
	DefaultAuthorMode string
	Sections          []SectionSpec
}

var specs = []Spec{
	{
		ID:                "tz",
		Folder:            "tz",
		Title:             "Техническое задание",
		DefaultAuthorMode: "team",
		Sections: []SectionSpec{
			{FileName: "01-intro.typ", Title: "Введение", Prompt: "Сформулируй введение и краткую характеристику области применения проекта."},
			{FileName: "02-basis.typ", Title: "Основания для разработки", Prompt: "Опиши основания для разработки, тему и исходные документы."},
			{FileName: "03-purpose.typ", Title: "Назначение разработки", Prompt: "Опиши функциональное и эксплуатационное назначение продукта."},
			{FileName: "04-requirements.typ", Title: "Требования к программе", Prompt: "Опиши требования к функциональности, надежности, условиям эксплуатации и совместимости."},
			{FileName: "05-documentation.typ", Title: "Требования к программной документации", Prompt: "Опиши состав и требования к программной документации."},
			{FileName: "06-economic.typ", Title: "Технико-экономические показатели", Prompt: "Опиши ориентировочные преимущества и экономический эффект."},
			{FileName: "07-stages.typ", Title: "Стадии и этапы разработки", Prompt: "Опиши стадии разработки и состав работ."},
			{FileName: "08-control.typ", Title: "Порядок контроля и приемки", Prompt: "Опиши контроль, приемку и условия сдачи."},
			{FileName: "09-references.typ", Title: "Список используемой литературы", Prompt: "Собери список нормативных документов и источников без вымышленных ссылок."},
			{FileName: "10-appendix.typ", Title: "Приложение", Prompt: "Подготовь приложения только если есть фактический материал; иначе оставь TODO."},
		},
	},
	{
		ID:                "pz",
		Folder:            "pz",
		Title:             "Пояснительная записка",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-overview.typ", Title: "Объект испытаний и назначение", Prompt: "Сделай раздел об объекте испытаний, наименовании программы и области применения."},
			{FileName: "02-requirements.typ", Title: "Требования к программе", Prompt: "Сформируй требования к программе на основе ТЗ, но без копирования формулировок."},
			{FileName: "03-tech.typ", Title: "Технические характеристики", Prompt: "Опиши постановку задачи, архитектуру, входные и выходные данные, интерфейс и стек."},
			{FileName: "04-economic.typ", Title: "Ожидаемые технико-экономические показатели", Prompt: "Опиши ожидаемый эффект, преимущества и ограничения."},
			{FileName: "05-sources.typ", Title: "Список литературы", Prompt: "Собери список использованных нормативных и технических источников без выдуманных материалов."},
		},
	},
	{
		ID:                "pmi",
		Folder:            "pmi",
		Title:             "Программа и методика испытаний",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-object.typ", Title: "Объект испытаний", Prompt: "Опиши объект испытаний, наименование программы и область применения."},
			{FileName: "02-goals.typ", Title: "Цель испытаний", Prompt: "Сформулируй цель испытаний и проверяемые свойства продукта."},
			{FileName: "03-program.typ", Title: "Требования к программе", Prompt: "Опиши функциональные требования, входные и выходные данные, интерфейс, надежность и совместимость."},
			{FileName: "04-docs.typ", Title: "Требования к программной документации", Prompt: "Опиши комплект документации, который должен быть предъявлен на испытаниях."},
			{FileName: "05-procedure.typ", Title: "Средства и порядок испытаний", Prompt: "Опиши технические средства, ПО и порядок проведения испытаний."},
			{FileName: "06-methods.typ", Title: "Методы испытаний", Prompt: "Опиши методики проверки ключевых требований и критерии успешного прохождения."},
			{FileName: "07-sources.typ", Title: "Список литературы", Prompt: "Собери релевантные источники и нормативные документы без выдуманных ссылок."},
		},
	},
	{
		ID:                "pmi-team",
		Folder:            "pmi-team",
		Title:             "Программа и методика испытаний",
		DefaultAuthorMode: "team",
		Sections: []SectionSpec{
			{FileName: "01-object.typ", Title: "Объект испытаний", Prompt: "Опиши объект испытаний и область применения командного проекта."},
			{FileName: "02-goals.typ", Title: "Цель испытаний", Prompt: "Сформулируй цели испытаний и зону приемки командной системы."},
			{FileName: "03-program.typ", Title: "Требования к программе", Prompt: "Опиши требования к клиентским приложениям, серверной части и взаимодействию компонентов."},
			{FileName: "04-docs.typ", Title: "Требования к программной документации", Prompt: "Опиши комплект документации для приемки командного проекта."},
			{FileName: "05-procedure.typ", Title: "Средства и порядок испытаний", Prompt: "Опиши стенд, ПО и порядок проведения испытаний для всей системы."},
			{FileName: "06-methods.typ", Title: "Методы испытаний", Prompt: "Опиши методы проверки интеграции, интерфейса и ключевых пользовательских сценариев."},
			{FileName: "07-sources.typ", Title: "Список литературы", Prompt: "Собери релевантные источники без выдуманных ссылок."},
		},
	},
	{
		ID:                "ro",
		Folder:            "ro",
		Title:             "Руководство оператора",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-purpose.typ", Title: "Назначение программы", Prompt: "Опиши функциональное и эксплуатационное назначение программы."},
			{FileName: "02-conditions.typ", Title: "Условия выполнения программы", Prompt: "Опиши аппаратные и программные требования и требования к пользователю."},
			{FileName: "03-run.typ", Title: "Выполнение программы", Prompt: "Опиши установку, запуск, базовые сценарии работы и остановку программы."},
			{FileName: "04-messages.typ", Title: "Сообщения оператору", Prompt: "Опиши сообщения пользователю и действия в исключительных ситуациях."},
		},
	},
}

func All() []Spec {
	return slices.Clone(specs)
}

func ByID(id string) (Spec, bool) {
	for _, spec := range specs {
		if spec.ID == id {
			return spec, true
		}
	}
	return Spec{}, false
}

func EnabledByMode(mode string) []Spec {
	var out []Spec
	for _, spec := range specs {
		if spec.ID == "pmi-team" && mode != "team" {
			continue
		}
		out = append(out, spec)
	}
	return out
}

func ResolveSelection(docArg, mode string) ([]Spec, error) {
	mode = strings.TrimSpace(mode)
	if mode != "single" && mode != "team" {
		return nil, fmt.Errorf("unsupported mode %q, expected single or team", mode)
	}

	docArg = strings.TrimSpace(docArg)
	if docArg == "" || docArg == "all" {
		return EnabledByMode(mode), nil
	}

	rawIDs := strings.Split(docArg, ",")
	out := make([]Spec, 0, len(rawIDs))
	seen := map[string]bool{}
	for _, rawID := range rawIDs {
		id := strings.TrimSpace(rawID)
		spec, ok := ByID(id)
		if !ok {
			return nil, fmt.Errorf("unknown document %q", id)
		}
		if seen[id] {
			continue
		}
		out = append(out, spec)
		seen[id] = true
	}

	return out, nil
}
