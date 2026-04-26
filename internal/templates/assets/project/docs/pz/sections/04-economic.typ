#import "../../../shared/typst/core.typ": paragraph, project-name, project-name-english, project-summary, project-type

= ОЖИДАЕМЫЕ ТЕХНИКО-ЭКОНОМИЧЕСКИЕ ПОКАЗАТЕЛИ

#paragraph[
  TODO: Обоснуйте полезность решения, ожидаемый эффект от внедрения и преимущества перед альтернативами. Нужные пункты:
  ==	Ориентировочная экономическая эффективность
  == Предполагаемая потребность
  == Экономические преимущества разработки по сравнению с отечественными и зарубежными аналогами

  Пример таблицы сравнения (ЭТО ТОЛЬКО ШАБЛОН):
  #let column_names = (
    [Яндекс Музыка],
    [Spotify],
    [ВК\ Музыка],
    [YouTube Music],
    [Apple Music],
    [*Our Project*],
  )

  #let plus = table.cell(
    fill: green.lighten(60%),
  )[+]

  #let minus = table.cell(
    fill: red.lighten(60%),
  )[-]

  #figure(
      caption: [Сравнение функциональных характеристик со стриминговыми сервисами],
      table(
          columns: (6cm,) + (2cm,) * column_names.len(),
          rows: (3cm, 1.5cm),
          align: center + horizon,
          table.header(
              [Функция],
              ..column_names.map(col => rotate(0deg, reflow: true, col))
          ),

          [Скачивание треков\ на устройство],   minus, 
          minus, minus, minus, minus, plus,
          [Нет рекламы], minus, minus, minus, minus, plus, plus,
          [Бесплатное прослушивание\ без ограничений], minus, minus, minus, minus, minus, plus,
          [Возможность использования в России], plus, minus, plus, minus, plus, plus,
          [Добавление своих треков], plus, minus, plus, minus, plus, plus,

          [*Итого*], [*2*], [*0*], [*2*], [*0*], [*3*], [*5*],
      )
  )
]
