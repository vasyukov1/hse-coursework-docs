# hse-coursework-docs

> Если есть вопросы или предложения, то пишите tg: [@overmindv](https://t.me/overmindv) / [@vasyukov_al ](https://t.me/vasyukov_al) 
> Можете улучшить этот проект и сделать запрос на PR

❗️ ВАЖНО: сейчас генерация не суперская, поэтому  я бы назвал этот инструмент заполнятором данных, которые легче править, чем писать всё с нуля.

Пример заполнения документов есть здесь: [Github](https://github.com/vasyukov1/HSE-FCS-SE-2-year/tree/main/Term-Paper).

---

`hse-coursework-docs` — CLI-инструмент для подготовки курсовой документации НИУ ВШЭ в формате Typst с последующей сборкой PDF.

Он помогает быстро подготовить:

- `ПЗ` — пояснительную записку
- `ПМИ` — программу и методику испытаний
- `ПМИ командное` — если есть командное ТЗ
- `РО` — руководство оператора

Генерация идёт по одному документу за запуск.

## Что нужно подготовить

Перед генерацией у вас должны быть:

- основное `ТЗ` в формате Typst
- при необходимости `командное ТЗ` для файла `ПМИ командное`
- один или несколько архивов с кодом или каталоги с репозиториями
- необязательный файл с заметками

## 🛠 Установка

### Через Homebrew

```bash
brew tap vasyukov1/hse-coursework-docs
brew install term-paper
brew install typst
```

Проверка:

```bash
term-paper help
typst --version
```

### Локальная сборка

```bash
git clone https://github.com/vasyukov1/hse-coursework-docs.git
cd hse-coursework-docs
go build -o term-paper ./cmd/term-paper
brew install typst
```

Чтобы запускать `term-paper` из любого каталога, добавьте бинарник в `PATH`:

```bash
sudo mkdir -p /usr/local/bin
sudo cp ./term-paper /usr/local/bin/term-paper
term-paper help
```

## Быстрый старт

### 1. Создайте проект

```bash
term-paper init
```

После этого появятся:

- `term-paper.yaml`
- `input/code/`
- `input/tz/`
- `input/tz-team/`
- `input/notes.txt`

### 2. Заполните `term-paper.yaml`

Пример:
```yaml
project:
  name: "Сервис для мониторинга влияния AI-треков Павла Дурова на рост TON"
  english_name: "TON Tracker via Pavel Durov's AI tracks"
  code: "05.01"
  summary: "Сервис, который отслеживает курс TON, анализирует влияние AI-треков Павла Дурова и показывает отчеты пользователю."
  type: "backend веб-сервиса"

student:
  name: "Имбулька Наталия Сергеевна"
  group: "БПИ285"

supervisor:
  name: "Дуров Павел Валерьевич"
  position: "CEO Telegram"

approver:
  name: "Н. А. Павлочев"
  position: "Академический руководитель образовательной программы \"Программная инженерия\", старший преподаватель департамента программной инженерии"

organization:
  name: "Национальный исследовательский университет \"Высшая школа экономики\""

inputs:
  code_dir: "./input/code"
  tz_dir: "./input/tz"
  notes_file: "./input/notes.txt"
  source_priority: "balanced"

ai:
  provider: "openrouter"
  base_url: "https://openrouter.ai/api/v1"
  model: "deepseek/deepseek-chat-v3-0324"
  api_key: ""
```

`source_priority`:

- `tz` — ТЗ важнее кода
- `code` — код важнее ТЗ
- `balanced` — оба источника одинаково важны

### 3. Положите входные файлы

#### Основное ТЗ

Положите Typst-файлы ТЗ в каталог:

```text
input/tz/
```

Лучший вариант:

- внутри `input/tz` есть `main.typ`
- `main.typ` включает остальные секции через `#include`

Если ТЗ уже разбито на несколько файлов, этого достаточно: генератор сам прочитает весь каталог через `main.typ`.

#### Командное ТЗ

Если нужен `ПМИ командное`, положите командное ТЗ в:

```text
input/tz-team/
```

#### Код проекта

Положите в `input/code/`:

- `.zip` архивы
- каталоги репозиториев
- несколько архивов сразу, если надо
- несколько репозиториев сразу, если надо

Примеры:

```text
input/code/backend.zip
input/code/frontend.zip
input/code/ml-service/
```

#### Заметки

В `input/notes.txt` можно написать:

- какие части ТЗ уже реализованы
- что расходится между ТЗ и кодом
- какие скриншоты потом надо вставить
- что особенно важно отразить в документации

Пример:

```text
Есть экран авторизации, экран списка ботов и экран просмотра отчета.
Нужно оставить TODO для скрина экрана авторизации.
В коде есть REST API и PostgreSQL, это важно показать в ПЗ.
```

## Если ТЗ разбито на секции

Чтобы собрать split-Typst в один файл:

```bash
term-paper bundle-typst --input ./input/tz --output ./input/tz-bundled.typ
```

Это удобно, если вы хотите:

- переиспользовать ТЗ как один файл
- отправить ТЗ отдельно
- вручную посмотреть, что именно пойдёт в генерацию

## Генерация документов

### ПЗ

```bash
term-paper generate-doc --doc pz
```

### ПМИ

```bash
term-paper generate-doc --doc pmi
```

### РО

```bash
term-paper generate-doc --doc ro
```

### ПМИ командное

```bash
term-paper generate-doc --doc pmi-team
```

## 😬 Если нужен только Typst-каркас без ИИ:

```bash
term-paper generate-doc --doc pz --skip-ai
```

По умолчанию результат генерации сразу записывается в рабочие секции:

```text
docs/<document>/sections/
```

Если нужен отдельный черновик без перезаписи рабочих секций:

```bash
term-paper generate-doc --doc pz --draft
```

Во время генерации в терминале выводятся логи по секциям.

## Улучшение существующего документа

```bash
term-paper improve-doc \
  --file docs/pz/sections/03-tech.typ \
  --prompt "сделай раздел подробнее, но не добавляй вымышленных технологий"
```

По умолчанию результат записывается в:

```text
docs/pz/sections/03-tech.typ.improved.typ
```

Чтобы переписать исходный файл:

```bash
term-paper improve-doc \
  --file docs/pz/sections/03-tech.typ \
  --prompt "сделай раздел подробнее" \
  --apply
```

## Сборка PDF

Собрать все уже созданные документы:

```bash
term-paper create-pdf
```

Собрать один документ:

```bash
term-paper create-pdf --doc pz
```

Результат попадает в `build/`.

Для быстрого цикла правок:

```bash
term-paper create-pdf --doc pz --watch
```

## Проверка проекта

```bash
term-paper doctor
```

Проверяется:

- установлен ли `typst`
- найден ли `term-paper.yaml`
- валиден ли конфиг
- существуют ли `input/tz`, `input/code`, `input/notes.txt`
- не повреждены ли уже созданные документы

## Справка по командам

```bash
term-paper help
term-paper help init
term-paper help generate-doc
term-paper help create-pdf
term-paper help bundle-typst
term-paper help improve-doc
```

## Какой ИИ подключать

Поддерживаются:

- `openrouter`
- `openai`
- `anthropic`
- `google`

### Рекомендуемые модели

Если нужен большой контекст и адекватная генерация документации:

- бесплатные или условно дешёвые: `deepseek/deepseek-chat-v3-0324`, `qwen/qwen3-235b-a22b`, `google/gemini-2.5-flash-preview`
- качественные платные: `anthropic/claude-sonnet-4`, `openai/gpt-5`, `google/gemini-2.5-pro`

Для очень больших архивов с кодом обычно лучше:

- класть в `input/code` только нужные репозитории и архивы
- дописывать важный контекст в `input/notes.txt`

<details>
<summary>Как получить API key в OpenRouter</summary>

1. Зайдите на [OpenRouter](https://openrouter.ai/).
2. Создайте аккаунт.
3. Откройте раздел `Keys`.
4. Создайте новый API key.
5. Вставьте его в `ai.api_key`.

Пример:

```yaml
ai:
  provider: "openrouter"
  base_url: "https://openrouter.ai/api/v1"
  model: "deepseek/deepseek-chat-v3-0324"
  api_key: "sk-or-..."
```
</details>

<details>
<summary>Как подключить OpenAI API</summary>

Нужен именно API key, а не просто подписка на сайт ChatGPT.

```yaml
ai:
  provider: "openai"
  base_url: "https://api.openai.com/v1"
  model: "gpt-5"
  api_key: "sk-..."
```
</details>

<details>
<summary>Как подключить Anthropic API</summary>

Нужен именно API key Anthropic, а не просто доступ к веб-версии Claude.

```yaml
ai:
  provider: "anthropic"
  base_url: "https://api.anthropic.com/v1"
  model: "claude-sonnet-4-20250514"
  api_key: "sk-ant-..."
```
</details>

## Что важно понимать

- `ПЗ` обычно самый важный документ для проверки и антиплагиата.
- `ПМИ командное` использует и основное ТЗ, и командное ТЗ.
- Базовая литература уже зашита в шаблоны и не должна ломаться.
- Если чего-то нет в ТЗ, коде или заметках, генератор должен оставлять `TODO`, а не выдумывать факты.
- Генератор пишет больше текста, чем обычно нужно в финальной версии: лишнее проще удалить, чем дописывать пустые разделы с нуля.
- Заголовки и порядок пунктов берутся из Typst-шаблонов в `docs/<doc>/sections`; их лучше не удалять вручную, иначе следующая генерация будет хуже держать структуру.

## Как улучшить качество генерации

- Заполните `project.summary` не одним предложением, а 5-10 предложениями: кто пользователь, какую задачу решает система, какие есть роли, модули, интерфейсы, данные и ограничения.
- В `input/notes.txt` перечислите реальные экраны, API-методы, таблицы БД, внешние сервисы, сценарии ошибок, ограничения и TODO для скриншотов.
- В `input/code` кладите не только корень проекта, но и важные README, OpenAPI/Swagger, docker-compose, миграции БД и конфиги окружения.
- Если генерация получилась слишком общей, допишите факты в `input/notes.txt` и запустите `term-paper improve-doc --file docs/<doc>/sections/<file>.typ --prompt "..." --apply`.
- Для ПЗ особенно полезно заранее перечислить аналоги и отличия от них: генератор сможет заполнить таблицы и экономический раздел без выдумывания конкурентов.
