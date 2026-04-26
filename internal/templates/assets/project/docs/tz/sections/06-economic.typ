#import "../../../shared/typst/core.typ": paragraph, project-name, project-name-english, project-summary, project-type

= ТЕХНИКО-ЭКОНОМИЧЕСКИЕ ПОКАЗАТЕЛИ

== Предполагаемая потребность

#paragraph[
  В этом разделе фиксируются ожидаемые преимущества разработки по сравнению с ручным трудом, существующими процессами или альтернативными решениями.
]

== Целевая аудитория

#paragraph[
  TODO: Опишите группы пользователей, для которых предназначена программа, и задачи, которые они решают с ее помощью.
]

== Преимущества перед аналогами

1. TODO: Опишите, какие ресурсы экономит проект.
2. TODO: Опишите качественные преимущества: скорость, точность, удобство, масштабируемость.
3. TODO: Если есть численные оценки, добавьте их с пояснениями.

ПРИМЕР СРАВНИТЕЛЬНОЙ ТАБЛИЦЫ:

#let column_names = (
  [Яндекс Музыка],
  [Spotify],
  [Apple Music],
)

#let plus = table.cell(
  fill: green.lighten(60%),
)[+]

#let minus = table.cell(
  fill: red.lighten(60%),
)[-]

#figure(
  caption: [Сравнение функциональных характеристик],
  table(
    columns: (6cm,) + (2cm,) * column_names.len(),
    rows: (3cm, 1.5cm),
    align: center + horizon,
    table.header(
      [Функция],
      ..column_names.map(col => rotate(0deg, reflow: true, col)),
    ),
    
    [Feature 1], minus, plus, minus,
    [Feature 2], plus, plus, plus,
    [Feature 3], minus, plus, minus,
    
    [*Итого*], [*2*], [*3*], [*5*],
  ),
)
