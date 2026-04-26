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
	Generate bool
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
			{FileName: "01-intro.typ", Title: "Введение"},
			{FileName: "02-basis.typ", Title: "Основания для разработки"},
			{FileName: "03-purpose.typ", Title: "Назначение разработки"},
			{FileName: "04-requirements.typ", Title: "Требования к программе"},
			{FileName: "05-documentation.typ", Title: "Требования к программной документации"},
			{FileName: "06-economic.typ", Title: "Технико-экономические показатели"},
			{FileName: "07-stages.typ", Title: "Стадии и этапы разработки"},
			{FileName: "08-control.typ", Title: "Порядок контроля и приемки"},
			{FileName: "09-references.typ", Title: "Список используемой литературы"},
			{FileName: "10-appendix.typ", Title: "Приложение"},
		},
	},
	{
		ID:                "pz",
		Folder:            "pz",
		Title:             "Пояснительная записка",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-overview.typ", Title: "Введение"},
			{FileName: "02-requirements.typ", Title: "Назначение и область применения", Prompt: "Опиши назначение программы и область применения. Нужен плотный технический текст без воды и без выдуманных деталей.", Generate: true},
			{FileName: "03-tech.typ", Title: "Технические характеристики", Prompt: "Подготовь основной технический раздел ПЗ: постановка задачи, описание решения, входные и выходные данные, архитектурные решения, стек, ограничения, состав модулей и особенности реализации. Это один из самых важных разделов документа.", Generate: true},
			{FileName: "04-economic.typ", Title: "Ожидаемые технико-экономические показатели", Prompt: "Опиши практическую полезность, преимущества и ожидаемый эффект. Не придумывай цифры, если их нет.", Generate: true},
			{FileName: "05-sources.typ", Title: "Список литературы"},
		},
	},
	{
		ID:                "pmi",
		Folder:            "pmi",
		Title:             "Программа и методика испытаний",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-object.typ", Title: "Объект испытаний"},
			{FileName: "02-goals.typ", Title: "Цель испытаний", Prompt: "Сформулируй цель испытаний и проверяемые свойства программы в официальном стиле. Нужно явно связать испытания с требованиями из ТЗ.", Generate: true},
			{FileName: "03-program.typ", Title: "Требования к программе", Prompt: "Подготовь основной раздел ПМИ: состав функций, входные и выходные данные, интерфейс, надежность, совместимость и другие проверяемые требования. Не добавляй вымышленных требований. Описывай только то, что действительно следует проверять.", Generate: true},
			{FileName: "04-docs.typ", Title: "Требования к программной документации"},
			{FileName: "05-procedure.typ", Title: "Средства и порядок испытаний", Prompt: "Опиши средства испытаний, окружение и порядок проведения проверки. Опирайся на код, ТЗ и очевидный стек проекта. Не придумывай стенды и инструменты, которых нет.", Generate: true},
			{FileName: "06-methods.typ", Title: "Методы испытаний", Prompt: "Опиши методы испытаний и критерии проверки функциональности. Если в проекте нужны скриншоты, оставляй TODO-плейсхолдеры. Нужны осмысленные сценарии проверки, а не короткий список общих фраз.", Generate: true},
			{FileName: "07-sources.typ", Title: "Список литературы"},
		},
	},
	{
		ID:                "pmi-team",
		Folder:            "pmi-team",
		Title:             "Программа и методика испытаний",
		DefaultAuthorMode: "team",
		Sections: []SectionSpec{
			{FileName: "01-object.typ", Title: "Объект испытаний"},
			{FileName: "02-goals.typ", Title: "Цель испытаний", Prompt: "Сформулируй цели испытаний и границы приемки командной системы. Используй общее ТЗ и при наличии командное ТЗ.", Generate: true},
			{FileName: "03-program.typ", Title: "Требования к программе", Prompt: "Подготовь требования к программе для командного ПМИ: клиентские части, серверные части, интеграции и взаимодействие компонентов. Командное ТЗ здесь особенно важно. Не добавляй компоненты, которых нет в ТЗ и коде.", Generate: true},
			{FileName: "04-docs.typ", Title: "Требования к программной документации"},
			{FileName: "05-procedure.typ", Title: "Средства и порядок испытаний", Prompt: "Опиши среду, средства и порядок испытаний для командного проекта, не выдумывая лишних компонент.", Generate: true},
			{FileName: "06-methods.typ", Title: "Методы испытаний", Prompt: "Опиши методы проверки интеграции, интерфейса и сквозных сценариев для командной системы. Нужен более подробный текст, чем просто перечень тестов.", Generate: true},
			{FileName: "07-sources.typ", Title: "Список литературы"},
		},
	},
	{
		ID:                "ro",
		Folder:            "ro",
		Title:             "Руководство оператора",
		DefaultAuthorMode: "single",
		Sections: []SectionSpec{
			{FileName: "01-purpose.typ", Title: "Назначение программы", Prompt: "Опиши функциональное и эксплуатационное назначение программы официальным техническим языком. Используй терминологию из ТЗ и проекта.", Generate: true},
			{FileName: "02-conditions.typ", Title: "Условия выполнения программы", Prompt: "Опиши условия выполнения программы: минимальные аппаратные и программные требования, а также требования к пользователю. Не придумывай конфигурации и версии, которых нельзя обосновать.", Generate: true},
			{FileName: "03-run.typ", Title: "Выполнение программы", Prompt: "Опиши установку, запуск и базовый сценарий работы программы. Если нужны скриншоты, оставь конкретные TODO. Текст должен быть пошаговым и практичным.", Generate: true},
			{FileName: "04-messages.typ", Title: "Сообщения оператору", Prompt: "Опиши сообщения оператору и действия в исключительных ситуациях. Не выдумывай сообщения, которых нет в проекте. Если точный текст неизвестен, опиши тип ситуации и действие пользователя без фантазии.", Generate: true},
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

func ResolveSelection(docArg string) ([]Spec, error) {
	docArg = strings.TrimSpace(docArg)
	if docArg == "" || docArg == "all" {
		return All(), nil
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
