#let config = yaml("../../term-paper.yaml")

#let project = config.project
#let student = config.student
#let supervisor = config.supervisor
#let approver = config.approver
#let organization = config.organization

#let faculty-name = "Факультет компьютерных наук"
#let program-name = "Образовательная программа \"Программная инженерия\""
#let current-year = datetime.today().display("[year]")

#let project-name() = project.name
#let project-name-english() = project.english_name
#let project-summary() = project.summary
#let project-type() = project.type

#let document-title(id) = if id == "tz" {
  "Техническое задание"
} else if id == "pz" {
  "Пояснительная записка"
} else if id == "pmi" or id == "pmi-team" {
  "Программа и методика испытаний"
} else if id == "ro" {
  "Руководство оператора"
} else {
  "Документ"
}

#let document-code(id) = if id == "tz" {
  "RU.17701729." + project.code + " ТЗ 01-1"
} else if id == "pz" {
  "RU.17701729." + project.code + " 81 01-1"
} else if id == "pmi" or id == "pmi-team" {
  "RU.17701729." + project.code + " 51 01-1"
} else if id == "ro" {
  "RU.17701729." + project.code + " 34 01-1"
} else {
  "RU.17701729." + project.code
}

#let approval-code(id) = document-code(id) + "-ЛУ"

#let paragraph(body) = {
  h(2em)
  body
}

#let render-document(id, body) = {
  set text(lang: "ru", size: 12pt, font: "Times New Roman")
  set par(
    first-line-indent: (amount: 2em, all: true),
    justify: true,
    spacing: 0.65em,
    leading: 0.65em,
  )

  set page(
    margin: (
      top: 22mm,
      left: 20mm,
      right: 12mm,
      bottom: 20mm,
    ),
    header: align(center, text(weight: "bold")[#document-code(id)]),
    footer: align(center, context counter(page).display()),
  )

  align(center)[
    #set text(weight: "bold")
    #organization.name
    #faculty-name
    #program-name

    #v(22mm)
    СОГЛАСОВАНО
    #supervisor.position
    #linebreak()
    #supervisor.name

    #v(10mm)
    УТВЕРЖДЕНО
    #approver.position
    #linebreak()
    #approver.name

    #v(20mm)
    #project-name()
    #document-title(id)
    ЛИСТ УТВЕРЖДЕНИЯ
    #approval-code(id)

    #v(18mm)
    Студент группы #student.group
    #linebreak()
    #student.name

    #v(18mm)
    #current-year
  ]

  pagebreak()

  align(center)[
    #set text(weight: "bold")
    УТВЕРЖДЕН
    #approval-code(id)

    #v(30mm)
    #project-name()
    #document-title(id)
    #document-code(id)

    #v(25mm)
    #current-year
  ]

  pagebreak()
  align(center, text(weight: "bold")[СОДЕРЖАНИЕ])
  outline(title: none, indent: 5mm)

  pagebreak()
  body

  pagebreak()
  align(center, text(weight: "bold")[ЛИСТ РЕГИСТРАЦИИ ИЗМЕНЕНИЙ])
  v(8mm)
  table(
    columns: (10mm,) + (15mm,) * 4 + (20mm,) * 2 + (auto,) + (auto,) * 2,
    rows: (auto,) * 3 + (9.5mm,) * 12,
    table.cell(colspan: 10, align: horizon, [Лист регистрации изменений]),
    table.cell(colspan: 5, align: horizon, [Номера листов (страниц)]),
    table.cell(rowspan: 2, align: horizon, [Всего листов (страниц в докум.)]),
    table.cell(rowspan: 2, align: horizon, [№ документа]),
    table.cell(rowspan: 2, align: horizon, [Входящий № сопроводительного докум. и дата]),
    table.cell(rowspan: 2, align: horizon, [Подп.]),
    table.cell(rowspan: 2, align: horizon, [Дата]),
    table.cell(align: horizon, rotate(-90deg, reflow: true)[Изм.]),
    table.cell(align: horizon, rotate(-90deg, reflow: true)[Измененных]),
    table.cell(align: horizon, rotate(-90deg, reflow: true)[Замененных]),
    table.cell(align: horizon, rotate(-90deg, reflow: true)[Новых]),
    table.cell(align: horizon, rotate(-90deg, reflow: true)[Аннулированных]),
  )
}
