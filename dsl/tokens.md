# Tokens

Here's a **complete token list** for the WXX DSL grammar we defined, including:

* Token name (symbol used in the grammar)
* Token type (for your Go `Token` struct)
* Example lexemes (what users would actually write)
* Notes (usage, edge cases)

---

### 📜 **WXX DSL Token List**

| Token Name | Type | Example(s) | Notes                   |
| ---------- | ---- | ---------- | ----------------------- |
| `SEMI`     | `;`  | `;`        | Statement terminator    |
| `ASSIGN`   | `:=` | `:=`       | Pascal-style assignment |
| `DOT`      | `.`  | `.`        | Member access           |
| `COMMA`    | `,`  | `,`        | Argument separator      |
| `LPAREN`   | `(`  | `(`        | Grouping, function call |
| `RPAREN`   | `)`  | `)`        | Closing grouping/call   |
| `LBRACKET` | `[`  | `[`        | Indexing into arrays    |
| `RBRACKET` | `]`  | `]`        | End of index            |

---

### 🔤 **Keywords** (case-insensitive if you want Pascal behavior)

| Token Name | Type    | Example(s) | Notes                               |
| ---------- | ------- | ---------- | ----------------------------------- |
| `IF`       | `if`    | `if`       | Control flow                        |
| `THEN`     | `then`  | `then`     | Required after `if` condition       |
| `ELSE`     | `else`  | `else`     | Optional branch                     |
| `END`      | `end`   | `end`      | Ends control blocks                 |
| `FOR`      | `for`   | `for`      | Loop keyword                        |
| `IN`       | `in`    | `in`       | Used in for-loops (`for x in expr`) |
| `DO`       | `do`    | `do`       | Begins the loop body                |
| `TRUE`     | `true`  | `true`     | Boolean literal                     |
| `FALSE`    | `false` | `false`    | Boolean literal                     |

---

### 📦 **Identifiers and Literals**

| Token Name   | Type          | Example(s)                      | Notes                                                         |
| ------------ | ------------- | ------------------------------- | ------------------------------------------------------------- |
| `IDENTIFIER` | string        | `map`, `hex`, `save`, `terrain` | Variable/function/property names                              |
| `NUMBER`     | float64 / int | `42`, `3.14`, `-7`              | Whole or decimal; distinguish int/float in parser if needed   |
| `STRING`     | string        | `"hello"`, `'swamp'`            | Allow double or single quotes (recommend consistent escaping) |

---

### 🔧 **Operators (BINOP)**

Handled as a single token type `BINOP`, with a value string like `+`, `-`, `*`, `/`, `=`, `<`, `>`, etc.

| Token Name | Type   | Example(s)                                          | Notes                                                       |
| ---------- | ------ | --------------------------------------------------- | ----------------------------------------------------------- |
| `BINOP`    | string | `+`, `-`, `*`, `/`, `=`, `<>`, `<`, `>`, `<=`, `>=` | You may lex these into a common type and dispatch on value. |

---

### 🧱 **Token Struct in Go**

You might define your `Token` type like this:

```go
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenAssign
	TokenSemicolon
	TokenDot
	TokenComma
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenIf
	TokenThen
	TokenElse
	TokenEnd
	TokenFor
	TokenIn
	TokenDo
	TokenTrue
	TokenFalse
	TokenBinOp
)

type Token struct {
	Type    TokenType
	Value   string      // raw string, e.g., "+", "map", "42"
	Line    int         // optional: for error messages
	Column  int         // optional: for error messages
}
```
