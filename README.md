# hse-coursework-docs

`hse-coursework-docs` — CLI-инструмент для подготовки курсовой документации НИУ ВШЭ в формате Typst с последующей сборкой PDF.

Он помогает быстро подготовить:

- `ПЗ` — пояснительную записку
- `ПМИ` — программу и методику испытаний
- `ПМИ командное` — если есть командное ТЗ
- `РО` — руководство оператора

Генерация идёт по одному документу за запуск. Это сделано специально, чтобы не тратить лишние токены и чтобы проще было контролировать результат.

## Что нужно подготовить

Перед генерацией у вас должны быть:

- основное `ТЗ` в формате Typst
- при необходимости `командное ТЗ` для `ПМИ командное`
- один или несколько архивов с кодом или каталоги с репозиториями
- необязательный файл с заметками

Пример темы проекта:

`Система мониторинга, насколько Павел Дуров одобрил бы ваш Telegram-бот на TON`

## Установка

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

```yaml
project:
  name: "Сервис оценки telegram-ботов по стандарту Дурова"
  english_name: "Telegram Bot Review Service by Durov Standard"
  code: "05.01-01"
  summary: "Сервис анализирует telegram-ботов, сохраняет результаты проверки и показывает статус соответствия внутренним критериям качества."
  type: "backend веб-сервиса"

student:
  name: "Иванов Иван Иванович"
  group: "БПИ228"

supervisor:
  name: "Петров Петр Петрович"
  position: "доцент департамента программной инженерии"

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
- несколько архивов сразу
- несколько репозиториев сразу

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

Если нужен только Typst-каркас без ИИ:

```bash
term-paper generate-doc --doc pz --skip-ai
```

По умолчанию результат генерации складывается в:

```text
docs/<document>/drafts/
```

Если хотите писать сразу в рабочие секции:

```bash
term-paper generate-doc --doc pz --apply
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
