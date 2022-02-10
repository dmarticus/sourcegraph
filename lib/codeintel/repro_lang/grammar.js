const { setupQunit } = require('@pollyjs/core')

module.exports = grammar({
  name: 'repro_lang',
  extras: $ => [/\s+/],
  word: $ => $.identifier,

  rules: {
    source_file: $ => repeat($._statement),
    _statement: $ => seq(choice($.definition_statement, $.reference_statement, $.comment), '\n'),
    definition_statement: $ =>
      seq(
        field('docstring', optional(seq($.docstring, '\n'))),
        'definition',
        field('name', $.identifier),
        field('roles', repeat($._definition_relations))
      ),
    reference_statement: $ => seq('reference', field('name', $.identifier)),
    _definition_relations: $ => choice($.implementation_relation, $.type_definition_relation, $.references_relation),
    implementation_relation: $ => seq('implementation', $.identifier),
    type_definition_relation: $ => seq('type_definition', $.identifier),
    references_relation: $ => seq('references', $.identifier),
    comment: $ => seq('#', /.*/),
    docstring: $ => seq('# docstring:', /.*/),
    identifier: $ => /[^\s]+/,
  },
})
