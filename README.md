# term-paper

`term-paper` — это CLI-инструмент для подготовки документации к курсовой работе в формате Typst.

Он делает две вещи:

1. Создаёт структуру документов и шаблоны.
2. Генерирует AI-черновики по вашему `ТЗ`, опциональному `ТЗ командному` и коду проекта.

Поддерживаемые документы:

- `ТЗ`
- `ПЗ`
- `ПМИ`
- `ПМИ командное`
- `РО`

Важно:

- командными считаются только `ТЗ` и `ПМИ`
- `ПЗ` считается самым важным документом и требует самой внимательной ручной вычитки
- если в документе нужны иллюстрации, генератор оставляет явные пометки вида `TODO: вставить скрин ...`

## Быстрый сценарий

1. Установить `term-paper`.
2. Установить `typst`.
3. Выполнить `term-paper init`.
4. Положить входные материалы в `input/`.
5. Заполнить `term-paper.yaml`.
6. Выполнить `term-paper generate`.
7. Проверить и поправить черновики.
8. Выполнить `term-paper create-pdf`.

## Установка

### Через Homebrew

```bash
brew tap vasyukov1/hse-coursework-docs
brew install term-paper
brew install typst
```

### Локальная сборка

```bash
go build -o term-paper ./cmd/term-paper
```

## Что нужно подготовить

Для лучшего результата стоит дать инструменту:

- основной `ТЗ` в `Typst`
- при наличии `ТЗ командное` тоже в `Typst`
- один или несколько архивов `.zip` с кодом
- и/или одну или несколько папок с локальными репозиториями
- дополнительные заметки по проекту
- список обязательных скриншотов, которые потом надо будет вставить вручную

Рекомендации:

- лучше давать `ТЗ` в `Typst`, а не в PDF
- если проект разбит на несколько репозиториев, лучше передать их все
- если AI не уверен, пусть лучше оставит `TODO`, чем выдумает текст

## Если ТЗ уже разбито на несколько Typst-файлов

Если ваше `ТЗ` состоит из `main.typ`, `body.typ`, секций и include-файлов, можно собрать его в один файл:

```bash
term-paper bundle-typst --input ./my-tz --output ./input/tz/bundled.typ
```

После этого в `term-paper.yaml` можно указать:

```yaml
sources:
  tz_path: ./input/tz/bundled.typ
```

Это удобно, если вы хотите много раз заново запускать генерацию по уже существующему `ТЗ`.

## Пример

Допустим, тема проекта такая:

`сервис для оценки роста TON после публикации AI-треков Павлом Дурова`

Запуск:

```bash
mkdir durov-ton-docs
cd durov-ton-docs
term-paper init --mode team
```

После этого можно заполнить `term-paper.yaml`, например так:

```yaml
project:
  name: Трекер TON посредством AI-треков Павла Дурова
  english_name: TON Tracker via Pavel Durov's AI tracks
  code: "05.01-01"
  summary: Сервис, который отслеживает курс TON, анализирует влияние AI-треков Павла Дурова и показывает отчеты пользователю.
  type: web-service
  mode: team

sources:
  tz_path: ./input/tz/main.typ
  team_tz_path: ./input/tz-team/main.typ
  code_paths:
    - ./input/code/backend.zip
    - ./input/code/frontend.zip
    - ../telegram-ton-bot
  notes:
    - ПЗ должно строго опираться на ТЗ и на реально реализованный код.
    - Если в коде и ТЗ есть расхождения, лучше оставить TODO, чем придумывать согласование.
    - Даже если тема шуточная, стиль документа должен оставаться серьёзным и техническим.
  screenshots:
    - TODO: вставить скрин экрана авторизации через Telegram
    - TODO: вставить скрин главного экрана со счётчиком TON
    - TODO: вставить скрин страницы отчёта

ai:
  provider: openrouter
  base_url: https://openrouter.ai/api/v1
  default_model: google/gemini-2.5-pro-preview
  api_key: sk-or-...
  api_key_env: OPENROUTER_API_KEY
```

Дальше:

```bash
term-paper generate
term-paper create-pdf
```

Результат:

- Typst-проект в `docs/`
- AI-черновики в `docs/<doc-id>/drafts/`
- PDF-файлы в `build/`

## Команды

### `term-paper init`

Создаёт:

- `term-paper.yaml`
- `input/tz/main.typ`
- `input/tz-team/main.typ`
- `input/code/`

Примеры:

```bash
term-paper init --mode single
term-paper init --mode team --output ./my-docs
```

### `term-paper bundle-typst`

Собирает Typst-источник из файла или каталога в один `.typ`.

Примеры:

```bash
term-paper bundle-typst --input ./docs/tz --output ./input/tz/bundled.typ
term-paper bundle-typst --input ./legacy-tz --entry main.typ --output ./input/tz/flat.typ
```

### `term-paper generate`

Читает `term-paper.yaml`, создаёт Typst-проект и, если есть AI-ключ и входные материалы, генерирует черновики для включённых документов.

Примеры:

```bash
term-paper generate
term-paper generate --doc pz
term-paper generate --skip-ai
```

Примечания:

- если `ai.api_key` пустой и соответствующая переменная окружения не задана, команда всё равно создаст структуру проекта, но пропустит AI-черновики
- если нужен только один документ, используйте `--doc`

### `term-paper doctor`

Проверяет:

- установлен ли `typst`
- валиден ли `term-paper.yaml`
- существуют ли пути из `sources.*`
- существуют ли все обязательные файлы документов

```bash
term-paper doctor
```

### `term-paper create-pdf`

Собирает PDF из Typst:

```bash
term-paper create-pdf
term-paper create-pdf --doc tz
term-paper create-pdf --doc pz --watch
```

### `term-paper ai draft`

Ручной перезапуск AI-генерации, если вы не хотите заново запускать весь `generate`:

```bash
term-paper ai draft --from tz
term-paper ai draft --from tz --doc pz
term-paper ai draft --from tz --doc pz --apply
```

По умолчанию черновики пишутся в:

```text
docs/<doc-id>/drafts/
```

Флаг `--apply` пишет сразу в:

```text
docs/<doc-id>/sections/
```

## Какие AI-провайдеры можно использовать

Сейчас поддерживаются:

- `openrouter`
- `openai`
- `anthropic`

### OpenRouter

Самый удобный вариант, если вы хотите быстро переключаться между Qwen, DeepSeek, Gemini, GPT и Claude через один API.

Пример:

```yaml
ai:
  provider: openrouter
  base_url: https://openrouter.ai/api/v1
  default_model: deepseek/deepseek-chat-v3-0324
  api_key_env: OPENROUTER_API_KEY
```

### OpenAI

Если вы хотите использовать модели OpenAI напрямую:

```yaml
ai:
  provider: openai
  base_url: https://api.openai.com/v1
  default_model: gpt-4.1
  api_key_env: OPENAI_API_KEY
```

Важно:

- обычная подписка ChatGPT и API OpenAI — это не одно и то же
- для работы через `term-paper` нужен именно API-ключ OpenAI

### Anthropic

Если вы хотите использовать Claude напрямую:

```yaml
ai:
  provider: anthropic
  base_url: https://api.anthropic.com/v1
  default_model: claude-sonnet-4-5
  api_key_env: ANTHROPIC_API_KEY
```

Важно:

- для работы нужен API-ключ Anthropic
- web-подписка Claude сама по себе не заменяет API-ключ

## Что именно AI использует при генерации

AI получает:

- основной `ТЗ`
- при необходимости `ТЗ командное`
- выдержки из кода и структуры проекта
- заметки пользователя
- список обязательных скриншотов
- стилевые примеры из ваших уже существующих документов

Стилевые примеры используются только как ориентир по тону и структуре. Факты берутся только из ваших входных материалов.

## Структура проекта после генерации

```text
term-paper.yaml
input/
  tz/
  tz-team/
  code/
shared/
  typst/
    core.typ
docs/
  tz/
  pz/
  pmi/
  pmi-team/
  ro/
build/
```

## Разработка локально

```bash
go test ./...
go build ./cmd/term-paper
```
