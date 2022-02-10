#include <tree_sitter/parser.h>

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 13
#define STATE_COUNT 28
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 23
#define ALIAS_COUNT 0
#define TOKEN_COUNT 11
#define EXTERNAL_TOKEN_COUNT 0
#define FIELD_COUNT 3
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 5

enum {
  sym_identifier = 1,
  anon_sym_LF = 2,
  anon_sym_definition = 3,
  anon_sym_reference = 4,
  anon_sym_implements = 5,
  anon_sym_typeDefines = 6,
  anon_sym_references = 7,
  anon_sym_POUND = 8,
  aux_sym_comment_token1 = 9,
  anon_sym_POUNDdocstring_COLON = 10,
  sym_source_file = 11,
  sym__statement = 12,
  sym_definition_statement = 13,
  sym_reference_statement = 14,
  sym__definition_relations = 15,
  sym_implementation_relation = 16,
  sym_type_definition_relation = 17,
  sym_references_relation = 18,
  sym_comment = 19,
  sym_docstring = 20,
  aux_sym_source_file_repeat1 = 21,
  aux_sym_definition_statement_repeat1 = 22,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_identifier] = "identifier",
  [anon_sym_LF] = "\n",
  [anon_sym_definition] = "definition",
  [anon_sym_reference] = "reference",
  [anon_sym_implements] = "implements",
  [anon_sym_typeDefines] = "typeDefines",
  [anon_sym_references] = "references",
  [anon_sym_POUND] = "#",
  [aux_sym_comment_token1] = "comment_token1",
  [anon_sym_POUNDdocstring_COLON] = "# docstring:",
  [sym_source_file] = "source_file",
  [sym__statement] = "_statement",
  [sym_definition_statement] = "definition_statement",
  [sym_reference_statement] = "reference_statement",
  [sym__definition_relations] = "_definition_relations",
  [sym_implementation_relation] = "implementation_relation",
  [sym_type_definition_relation] = "type_definition_relation",
  [sym_references_relation] = "references_relation",
  [sym_comment] = "comment",
  [sym_docstring] = "docstring",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_definition_statement_repeat1] = "definition_statement_repeat1",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_identifier] = sym_identifier,
  [anon_sym_LF] = anon_sym_LF,
  [anon_sym_definition] = anon_sym_definition,
  [anon_sym_reference] = anon_sym_reference,
  [anon_sym_implements] = anon_sym_implements,
  [anon_sym_typeDefines] = anon_sym_typeDefines,
  [anon_sym_references] = anon_sym_references,
  [anon_sym_POUND] = anon_sym_POUND,
  [aux_sym_comment_token1] = aux_sym_comment_token1,
  [anon_sym_POUNDdocstring_COLON] = anon_sym_POUNDdocstring_COLON,
  [sym_source_file] = sym_source_file,
  [sym__statement] = sym__statement,
  [sym_definition_statement] = sym_definition_statement,
  [sym_reference_statement] = sym_reference_statement,
  [sym__definition_relations] = sym__definition_relations,
  [sym_implementation_relation] = sym_implementation_relation,
  [sym_type_definition_relation] = sym_type_definition_relation,
  [sym_references_relation] = sym_references_relation,
  [sym_comment] = sym_comment,
  [sym_docstring] = sym_docstring,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_definition_statement_repeat1] = aux_sym_definition_statement_repeat1,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [sym_identifier] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_LF] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_definition] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_reference] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_implements] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_typeDefines] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_references] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_POUND] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_comment_token1] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_POUNDdocstring_COLON] = {
    .visible = true,
    .named = false,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym__statement] = {
    .visible = false,
    .named = true,
  },
  [sym_definition_statement] = {
    .visible = true,
    .named = true,
  },
  [sym_reference_statement] = {
    .visible = true,
    .named = true,
  },
  [sym__definition_relations] = {
    .visible = false,
    .named = true,
  },
  [sym_implementation_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_type_definition_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_references_relation] = {
    .visible = true,
    .named = true,
  },
  [sym_comment] = {
    .visible = true,
    .named = true,
  },
  [sym_docstring] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_definition_statement_repeat1] = {
    .visible = false,
    .named = false,
  },
};

enum {
  field_docstring = 1,
  field_name = 2,
  field_roles = 3,
};

static const char * const ts_field_names[] = {
  [0] = NULL,
  [field_docstring] = "docstring",
  [field_name] = "name",
  [field_roles] = "roles",
};

static const TSFieldMapSlice ts_field_map_slices[PRODUCTION_ID_COUNT] = {
  [1] = {.index = 0, .length = 1},
  [2] = {.index = 1, .length = 2},
  [3] = {.index = 3, .length = 3},
  [4] = {.index = 6, .length = 4},
};

static const TSFieldMapEntry ts_field_map_entries[] = {
  [0] =
    {field_name, 1},
  [1] =
    {field_name, 1},
    {field_roles, 2},
  [3] =
    {field_docstring, 0},
    {field_docstring, 1},
    {field_name, 3},
  [6] =
    {field_docstring, 0},
    {field_docstring, 1},
    {field_name, 3},
    {field_roles, 4},
};

static const TSSymbol ts_alias_sequences[PRODUCTION_ID_COUNT][MAX_ALIAS_SEQUENCE_LENGTH] = {
  [0] = {0},
};

static const uint16_t ts_non_terminal_alias_map[] = {
  0,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(32);
      if (lookahead == '#') ADVANCE(39);
      if (lookahead == 'd') ADVANCE(44);
      if (lookahead == 'r') ADVANCE(47);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 1:
      if (lookahead == '\n') ADVANCE(33);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(1)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 2:
      if (lookahead == '\n') ADVANCE(33);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(2)
      END_STATE();
    case 3:
      if (lookahead == ':') ADVANCE(42);
      END_STATE();
    case 4:
      if (lookahead == 'c') ADVANCE(27);
      END_STATE();
    case 5:
      if (lookahead == 'c') ADVANCE(8);
      END_STATE();
    case 6:
      if (lookahead == 'd') ADVANCE(23);
      END_STATE();
    case 7:
      if (lookahead == 'e') ADVANCE(13);
      END_STATE();
    case 8:
      if (lookahead == 'e') ADVANCE(36);
      END_STATE();
    case 9:
      if (lookahead == 'e') ADVANCE(26);
      END_STATE();
    case 10:
      if (lookahead == 'e') ADVANCE(12);
      END_STATE();
    case 11:
      if (lookahead == 'e') ADVANCE(21);
      END_STATE();
    case 12:
      if (lookahead == 'f') ADVANCE(9);
      END_STATE();
    case 13:
      if (lookahead == 'f') ADVANCE(18);
      END_STATE();
    case 14:
      if (lookahead == 'g') ADVANCE(3);
      END_STATE();
    case 15:
      if (lookahead == 'i') ADVANCE(19);
      END_STATE();
    case 16:
      if (lookahead == 'i') ADVANCE(24);
      END_STATE();
    case 17:
      if (lookahead == 'i') ADVANCE(29);
      END_STATE();
    case 18:
      if (lookahead == 'i') ADVANCE(22);
      END_STATE();
    case 19:
      if (lookahead == 'n') ADVANCE(14);
      END_STATE();
    case 20:
      if (lookahead == 'n') ADVANCE(34);
      END_STATE();
    case 21:
      if (lookahead == 'n') ADVANCE(5);
      END_STATE();
    case 22:
      if (lookahead == 'n') ADVANCE(17);
      END_STATE();
    case 23:
      if (lookahead == 'o') ADVANCE(4);
      END_STATE();
    case 24:
      if (lookahead == 'o') ADVANCE(20);
      END_STATE();
    case 25:
      if (lookahead == 'r') ADVANCE(15);
      END_STATE();
    case 26:
      if (lookahead == 'r') ADVANCE(11);
      END_STATE();
    case 27:
      if (lookahead == 's') ADVANCE(28);
      END_STATE();
    case 28:
      if (lookahead == 't') ADVANCE(25);
      END_STATE();
    case 29:
      if (lookahead == 't') ADVANCE(16);
      END_STATE();
    case 30:
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(30)
      if (lookahead != 0) ADVANCE(60);
      END_STATE();
    case 31:
      if (eof) ADVANCE(32);
      if (lookahead == '#') ADVANCE(38);
      if (lookahead == 'd') ADVANCE(7);
      if (lookahead == 'r') ADVANCE(10);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(31)
      END_STATE();
    case 32:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_LF);
      if (lookahead == '\n') ADVANCE(33);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(anon_sym_definition);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(anon_sym_definition);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_reference);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(anon_sym_reference);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 38:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(6);
      END_STATE();
    case 39:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead == ' ') ADVANCE(6);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r') ADVANCE(60);
      END_STATE();
    case 40:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(40);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(41);
      END_STATE();
    case 41:
      ACCEPT_TOKEN(aux_sym_comment_token1);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(41);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(anon_sym_POUNDdocstring_COLON);
      END_STATE();
    case 43:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'c') ADVANCE(46);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'e') ADVANCE(49);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'e') ADVANCE(58);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 46:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'e') ADVANCE(37);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'e') ADVANCE(50);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'e') ADVANCE(54);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 49:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'f') ADVANCE(51);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 50:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'f') ADVANCE(45);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'i') ADVANCE(56);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'i') ADVANCE(59);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'i') ADVANCE(57);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 54:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'n') ADVANCE(43);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'n') ADVANCE(35);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 56:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'n') ADVANCE(52);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'o') ADVANCE(55);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 58:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 'r') ADVANCE(48);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 59:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == 't') ADVANCE(53);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    case 60:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n' &&
          lookahead != '\r' &&
          lookahead != ' ') ADVANCE(60);
      END_STATE();
    default:
      return false;
  }
}

static bool ts_lex_keywords(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (lookahead == 'i') ADVANCE(1);
      if (lookahead == 'r') ADVANCE(2);
      if (lookahead == 't') ADVANCE(3);
      if (lookahead == '\t' ||
          lookahead == '\n' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0)
      END_STATE();
    case 1:
      if (lookahead == 'm') ADVANCE(4);
      END_STATE();
    case 2:
      if (lookahead == 'e') ADVANCE(5);
      END_STATE();
    case 3:
      if (lookahead == 'y') ADVANCE(6);
      END_STATE();
    case 4:
      if (lookahead == 'p') ADVANCE(7);
      END_STATE();
    case 5:
      if (lookahead == 'f') ADVANCE(8);
      END_STATE();
    case 6:
      if (lookahead == 'p') ADVANCE(9);
      END_STATE();
    case 7:
      if (lookahead == 'l') ADVANCE(10);
      END_STATE();
    case 8:
      if (lookahead == 'e') ADVANCE(11);
      END_STATE();
    case 9:
      if (lookahead == 'e') ADVANCE(12);
      END_STATE();
    case 10:
      if (lookahead == 'e') ADVANCE(13);
      END_STATE();
    case 11:
      if (lookahead == 'r') ADVANCE(14);
      END_STATE();
    case 12:
      if (lookahead == 'D') ADVANCE(15);
      END_STATE();
    case 13:
      if (lookahead == 'm') ADVANCE(16);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(17);
      END_STATE();
    case 15:
      if (lookahead == 'e') ADVANCE(18);
      END_STATE();
    case 16:
      if (lookahead == 'e') ADVANCE(19);
      END_STATE();
    case 17:
      if (lookahead == 'n') ADVANCE(20);
      END_STATE();
    case 18:
      if (lookahead == 'f') ADVANCE(21);
      END_STATE();
    case 19:
      if (lookahead == 'n') ADVANCE(22);
      END_STATE();
    case 20:
      if (lookahead == 'c') ADVANCE(23);
      END_STATE();
    case 21:
      if (lookahead == 'i') ADVANCE(24);
      END_STATE();
    case 22:
      if (lookahead == 't') ADVANCE(25);
      END_STATE();
    case 23:
      if (lookahead == 'e') ADVANCE(26);
      END_STATE();
    case 24:
      if (lookahead == 'n') ADVANCE(27);
      END_STATE();
    case 25:
      if (lookahead == 's') ADVANCE(28);
      END_STATE();
    case 26:
      if (lookahead == 's') ADVANCE(29);
      END_STATE();
    case 27:
      if (lookahead == 'e') ADVANCE(30);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(anon_sym_implements);
      END_STATE();
    case 29:
      ACCEPT_TOKEN(anon_sym_references);
      END_STATE();
    case 30:
      if (lookahead == 's') ADVANCE(31);
      END_STATE();
    case 31:
      ACCEPT_TOKEN(anon_sym_typeDefines);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0},
  [1] = {.lex_state = 31},
  [2] = {.lex_state = 31},
  [3] = {.lex_state = 31},
  [4] = {.lex_state = 1},
  [5] = {.lex_state = 1},
  [6] = {.lex_state = 1},
  [7] = {.lex_state = 1},
  [8] = {.lex_state = 1},
  [9] = {.lex_state = 31},
  [10] = {.lex_state = 1},
  [11] = {.lex_state = 1},
  [12] = {.lex_state = 1},
  [13] = {.lex_state = 2},
  [14] = {.lex_state = 2},
  [15] = {.lex_state = 31},
  [16] = {.lex_state = 30},
  [17] = {.lex_state = 30},
  [18] = {.lex_state = 30},
  [19] = {.lex_state = 30},
  [20] = {.lex_state = 2},
  [21] = {.lex_state = 30},
  [22] = {.lex_state = 2},
  [23] = {.lex_state = 2},
  [24] = {.lex_state = 0},
  [25] = {.lex_state = 40},
  [26] = {.lex_state = 40},
  [27] = {.lex_state = 30},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_identifier] = ACTIONS(1),
    [anon_sym_definition] = ACTIONS(1),
    [anon_sym_reference] = ACTIONS(1),
    [anon_sym_implements] = ACTIONS(1),
    [anon_sym_typeDefines] = ACTIONS(1),
    [anon_sym_references] = ACTIONS(1),
    [anon_sym_POUND] = ACTIONS(1),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(24),
    [sym__statement] = STATE(2),
    [sym_definition_statement] = STATE(13),
    [sym_reference_statement] = STATE(13),
    [sym_comment] = STATE(13),
    [sym_docstring] = STATE(23),
    [aux_sym_source_file_repeat1] = STATE(2),
    [ts_builtin_sym_end] = ACTIONS(3),
    [anon_sym_definition] = ACTIONS(5),
    [anon_sym_reference] = ACTIONS(7),
    [anon_sym_POUND] = ACTIONS(9),
    [anon_sym_POUNDdocstring_COLON] = ACTIONS(11),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 8,
    ACTIONS(5), 1,
      anon_sym_definition,
    ACTIONS(7), 1,
      anon_sym_reference,
    ACTIONS(9), 1,
      anon_sym_POUND,
    ACTIONS(11), 1,
      anon_sym_POUNDdocstring_COLON,
    ACTIONS(13), 1,
      ts_builtin_sym_end,
    STATE(23), 1,
      sym_docstring,
    STATE(3), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(13), 3,
      sym_definition_statement,
      sym_reference_statement,
      sym_comment,
  [28] = 8,
    ACTIONS(15), 1,
      ts_builtin_sym_end,
    ACTIONS(17), 1,
      anon_sym_definition,
    ACTIONS(20), 1,
      anon_sym_reference,
    ACTIONS(23), 1,
      anon_sym_POUND,
    ACTIONS(26), 1,
      anon_sym_POUNDdocstring_COLON,
    STATE(23), 1,
      sym_docstring,
    STATE(3), 2,
      sym__statement,
      aux_sym_source_file_repeat1,
    STATE(13), 3,
      sym_definition_statement,
      sym_reference_statement,
      sym_comment,
  [56] = 5,
    ACTIONS(29), 1,
      anon_sym_LF,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_typeDefines,
    ACTIONS(35), 1,
      anon_sym_references,
    STATE(6), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [76] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_typeDefines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(37), 1,
      anon_sym_LF,
    STATE(4), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [96] = 5,
    ACTIONS(39), 1,
      anon_sym_LF,
    ACTIONS(41), 1,
      anon_sym_implements,
    ACTIONS(44), 1,
      anon_sym_typeDefines,
    ACTIONS(47), 1,
      anon_sym_references,
    STATE(6), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [116] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_typeDefines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(50), 1,
      anon_sym_LF,
    STATE(8), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [136] = 5,
    ACTIONS(31), 1,
      anon_sym_implements,
    ACTIONS(33), 1,
      anon_sym_typeDefines,
    ACTIONS(35), 1,
      anon_sym_references,
    ACTIONS(52), 1,
      anon_sym_LF,
    STATE(6), 5,
      sym__definition_relations,
      sym_implementation_relation,
      sym_type_definition_relation,
      sym_references_relation,
      aux_sym_definition_statement_repeat1,
  [156] = 2,
    ACTIONS(56), 1,
      anon_sym_POUND,
    ACTIONS(54), 4,
      ts_builtin_sym_end,
      anon_sym_definition,
      anon_sym_reference,
      anon_sym_POUNDdocstring_COLON,
  [166] = 2,
    ACTIONS(58), 1,
      anon_sym_LF,
    ACTIONS(60), 3,
      anon_sym_implements,
      anon_sym_typeDefines,
      anon_sym_references,
  [175] = 2,
    ACTIONS(62), 1,
      anon_sym_LF,
    ACTIONS(64), 3,
      anon_sym_implements,
      anon_sym_typeDefines,
      anon_sym_references,
  [184] = 2,
    ACTIONS(66), 1,
      anon_sym_LF,
    ACTIONS(68), 3,
      anon_sym_implements,
      anon_sym_typeDefines,
      anon_sym_references,
  [193] = 1,
    ACTIONS(70), 1,
      anon_sym_LF,
  [197] = 1,
    ACTIONS(72), 1,
      anon_sym_LF,
  [201] = 1,
    ACTIONS(74), 1,
      anon_sym_definition,
  [205] = 1,
    ACTIONS(76), 1,
      sym_identifier,
  [209] = 1,
    ACTIONS(78), 1,
      sym_identifier,
  [213] = 1,
    ACTIONS(80), 1,
      sym_identifier,
  [217] = 1,
    ACTIONS(82), 1,
      sym_identifier,
  [221] = 1,
    ACTIONS(84), 1,
      anon_sym_LF,
  [225] = 1,
    ACTIONS(86), 1,
      sym_identifier,
  [229] = 1,
    ACTIONS(88), 1,
      anon_sym_LF,
  [233] = 1,
    ACTIONS(90), 1,
      anon_sym_LF,
  [237] = 1,
    ACTIONS(92), 1,
      ts_builtin_sym_end,
  [241] = 1,
    ACTIONS(94), 1,
      aux_sym_comment_token1,
  [245] = 1,
    ACTIONS(96), 1,
      aux_sym_comment_token1,
  [249] = 1,
    ACTIONS(98), 1,
      sym_identifier,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 28,
  [SMALL_STATE(4)] = 56,
  [SMALL_STATE(5)] = 76,
  [SMALL_STATE(6)] = 96,
  [SMALL_STATE(7)] = 116,
  [SMALL_STATE(8)] = 136,
  [SMALL_STATE(9)] = 156,
  [SMALL_STATE(10)] = 166,
  [SMALL_STATE(11)] = 175,
  [SMALL_STATE(12)] = 184,
  [SMALL_STATE(13)] = 193,
  [SMALL_STATE(14)] = 197,
  [SMALL_STATE(15)] = 201,
  [SMALL_STATE(16)] = 205,
  [SMALL_STATE(17)] = 209,
  [SMALL_STATE(18)] = 213,
  [SMALL_STATE(19)] = 217,
  [SMALL_STATE(20)] = 221,
  [SMALL_STATE(21)] = 225,
  [SMALL_STATE(22)] = 229,
  [SMALL_STATE(23)] = 233,
  [SMALL_STATE(24)] = 237,
  [SMALL_STATE(25)] = 241,
  [SMALL_STATE(26)] = 245,
  [SMALL_STATE(27)] = 249,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 0),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT(27),
  [9] = {.entry = {.count = 1, .reusable = false}}, SHIFT(26),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1),
  [15] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2),
  [17] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(16),
  [20] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(27),
  [23] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(26),
  [26] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2), SHIFT_REPEAT(25),
  [29] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 5, .production_id = 4),
  [31] = {.entry = {.count = 1, .reusable = false}}, SHIFT(17),
  [33] = {.entry = {.count = 1, .reusable = false}}, SHIFT(18),
  [35] = {.entry = {.count = 1, .reusable = false}}, SHIFT(19),
  [37] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 4, .production_id = 3),
  [39] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_definition_statement_repeat1, 2),
  [41] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(17),
  [44] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(18),
  [47] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_definition_statement_repeat1, 2), SHIFT_REPEAT(19),
  [50] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 2, .production_id = 1),
  [52] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_definition_statement, 3, .production_id = 2),
  [54] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym__statement, 2),
  [56] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym__statement, 2),
  [58] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_references_relation, 2),
  [60] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_references_relation, 2),
  [62] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_type_definition_relation, 2),
  [64] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_type_definition_relation, 2),
  [66] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_implementation_relation, 2),
  [68] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_implementation_relation, 2),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [72] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_docstring, 2),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
  [82] = {.entry = {.count = 1, .reusable = true}}, SHIFT(10),
  [84] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_comment, 2),
  [86] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [88] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_reference_statement, 2, .production_id = 1),
  [90] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [92] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [94] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [96] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [98] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
};

#ifdef __cplusplus
extern "C" {
#endif
#ifdef _WIN32
#define extern __declspec(dllexport)
#endif

extern const TSLanguage *tree_sitter_repro_lang(void) {
  static const TSLanguage language = {
    .version = LANGUAGE_VERSION,
    .symbol_count = SYMBOL_COUNT,
    .alias_count = ALIAS_COUNT,
    .token_count = TOKEN_COUNT,
    .external_token_count = EXTERNAL_TOKEN_COUNT,
    .state_count = STATE_COUNT,
    .large_state_count = LARGE_STATE_COUNT,
    .production_id_count = PRODUCTION_ID_COUNT,
    .field_count = FIELD_COUNT,
    .max_alias_sequence_length = MAX_ALIAS_SEQUENCE_LENGTH,
    .parse_table = &ts_parse_table[0][0],
    .small_parse_table = ts_small_parse_table,
    .small_parse_table_map = ts_small_parse_table_map,
    .parse_actions = ts_parse_actions,
    .symbol_names = ts_symbol_names,
    .field_names = ts_field_names,
    .field_map_slices = ts_field_map_slices,
    .field_map_entries = ts_field_map_entries,
    .symbol_metadata = ts_symbol_metadata,
    .public_symbol_map = ts_symbol_map,
    .alias_map = ts_non_terminal_alias_map,
    .alias_sequences = &ts_alias_sequences[0][0],
    .lex_modes = ts_lex_modes,
    .lex_fn = ts_lex,
    .keyword_lex_fn = ts_lex_keywords,
    .keyword_capture_token = sym_identifier,
  };
  return &language;
}
#ifdef __cplusplus
}
#endif
