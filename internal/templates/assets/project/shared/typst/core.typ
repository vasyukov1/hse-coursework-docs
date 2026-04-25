#let config = yaml("../../term-paper.yaml")

#let project = config.project
#let organization = config.organization
#let supervisors = config.supervisors
#let docs = config.docs

#let participants() = config.participants
#let primary-participant() = participants().at(0)
#let document(id) = docs.at(id)
#let document-title(id) = document(id).title
#let document-code(id) = document(id).code
#let document-annotation(id) = document(id).annotation
#let approval-code(id) = document-code(id) + "-ЛУ"
#let project-name() = project.name
#let project-name-english() = project.english_name
#let project-summary() = project.summary
#let project-type() = project.type
#let is-team-mode() = project.mode == "team" or participants().len() > 1
#let authors(id) = if document(id).author_mode == "team" { participants() } else { (primary-participant(),) }

#let paragraph(body) = {
  h(2em)
  body
}

#let render-document(id, body) = {
  let un(n) = "_" * n

  let storage-table = {
    set text(size: 10pt)
    place(
      bottom + left,
      dx: 5mm,
      dy: -10mm,
      rotate(
        -90deg,
        reflow: true,
        table(
          columns: (25mm, 35mm, 25mm, 25mm, 35mm),
          rows: (5mm, 7mm),
          align: center,
          [Инв.№ подп], [Подп. и дата], [Взам. инв.№], [Инв.№ дубл.], [Подп. и дата],
        ),
      ),
    )
  }

  let authors-block = authors(id).map(author => [
    Студент группы #author.group

    #un(13) / #author.name /

    "#un(3)" #un(15) #organization.year г.
  ]).join[#linebreak() #linebreak()]

  let approval-page = {
    page(
      header: none,
      footer: none,
      margin: (
        left: 20mm,
        right: 10mm,
        top: 25mm,
        bottom: 15mm,
      ),
      foreground: storage-table,
    )[
      #set align(center)
      #grid(
        columns: (1fr),
        row-gutter: 1fr,
        [
          #set par(spacing: 0.65em)
          #set text(weight: "bold")
          #organization.university
          #organization.faculty
          #organization.program
        ],
        grid(
          columns: (1fr, 1fr),
          inset: (x: 10mm, y: 3mm),
          align: center,
          [
            СОГЛАСОВАНО

            #supervisors.agreed_by.position
          ],
          [
            УТВЕРЖДЕНО

            #supervisors.approved_by.position
          ],
          [
            #un(13) #supervisors.agreed_by.name

            "#un(3)" #un(13) #organization.year г.
          ],
          [
            #un(13) #supervisors.approved_by.name

            "#un(3)" #un(13) #organization.year г.
          ],
        ),
        [
          #set text(size: 14pt, weight: "bold")
          #set par(spacing: 2em)
          #project-name()
          #document-title(id)
          ЛИСТ УТВЕРЖДЕНИЯ
          #approval-code(id)
        ],
        align(right)[
          #set par(spacing: 1em)
          #(if authors(id).len() > 1 [Исполнители:] else [Исполнитель:])
          #authors-block
        ],
        [
          #set text(weight: "bold")
          #organization.year
        ],
      )
    ]
    counter(page).update(1)
  }

  let title-page = {
    page(
      header: none,
      footer: none,
      margin: (
        left: 20mm,
        right: 10mm,
        top: 25mm,
        bottom: 15mm,
      ),
      foreground: storage-table,
    )[
      #set align(center)
      #grid(
        columns: (1fr),
        row-gutter: 1fr,
        [
          #set align(left)
          #set par(spacing: 2em)
          УТВЕРЖДЕН
          #approval-code(id)
        ],
        [
          #set text(size: 14pt, weight: "bold")
          #set par(spacing: 2em)
          #project-name()
          #document-title(id)
          #document-code(id)
          Листов #context { counter(page).final().at(0) }
        ],
        [
          #set text(weight: "bold")
          #organization.year
        ],
      )
    ]
  }

  let registration-page = {
    page(
      header: none,
      footer: none,
      margin: (
        left: 20mm,
        right: 10mm,
        top: 25mm,
        bottom: 15mm,
      ),
    )[
      #set align(center)
      #set text(size: 14pt, weight: "bold")
      ЛИСТ РЕГИСТРАЦИИ ИЗМЕНЕНИЙ
      #v(1fr)
      #table(
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
    ]
  }

  let outline-page = {
    pagebreak(weak: true)
    align(center, text(weight: "bold")[СОДЕРЖАНИЕ])
    outline(title: none, indent: 5mm)
  }

  set text(lang: "ru", size: 12pt, font: "Times New Roman")

  approval-page
  title-page

  set page(
    margin: (
      top: 25mm,
      left: 20mm,
      right: 10mm,
      bottom: 47mm,
    ),
    header: [
      #set align(center)
      #set text(weight: "bold")
      #context counter(page).display()
      #document-code(id)
    ],
    footer: [
      #table(
        columns: (2fr, 1fr, 1fr, 1fr, 1fr),
        align: center,
        rows: 7mm,
        [], [], [], [], [],
        [Изм.], [Лист], [№ докум.], [Подп.], [Дата],
        document-code(id), [], [], [], [],
        [Инв. № подл.], [Подп. и дата], [Взам. Инв. №], [Инв. № дубл.], [Подп. и дата],
      )
    ],
  )

  set par(
    first-line-indent: (amount: 2em, all: true),
    justify: true,
    spacing: 0.65em,
    leading: 0.65em,
  )

  set list(indent: 2em, spacing: 0.65em, marker: "-")
  set enum(indent: 2em, spacing: 0.65em)
  set heading(numbering: "1.")

  show heading.where(level: 1): h => {
    set align(center)
    set text(weight: "bold", size: 12pt)
    pagebreak(weak: true)
    if h.numbering != none [
      #counter(heading).display(h.numbering) #h.body
    ] else [
      #h.body
    ]
  }

  show heading.where(level: 2): h => {
    set text(weight: "bold", size: 12pt)
    block(inset: (left: 1em))[#counter(heading).display() #h.body]
  }

  show heading.where(level: 3): h => {
    set text(weight: "bold", size: 12pt)
    block(inset: (left: 3em))[#counter(heading).display() #h.body]
  }

  if document-annotation(id) != "" {
    pagebreak(weak: true)
    align(center, text(weight: "bold", size: 12pt, [АННОТАЦИЯ]))
    paragraph(document-annotation(id))
  }

  outline-page
  body
  registration-page
}
