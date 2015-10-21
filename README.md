# Char Language Specification
## General
Char is intended as a general-purpose language with a focus on readability and conciseness without resorting to a a large number of operators and special characters.

The core tenets guiding the syntax are as follows:
* There should be a single, clear way to do a task.
* An expression should read as a sentence.
* Keywords are preferred over operators if they are of similar length.
* Reducing typing time is important.
* Things that can be inferred by the compiler should be, unless it hampers readability.
* Special characters to help the compiler are a waste of typing time.
* Any place where you have to repeat a leading keyword should also allow an indented block.
* When switching a single statement to an indented block, you shouldn't have to rewrite the command itself.
### Characters
```
newline             = ? the Unicode code point U+000A ?
unicode_char        = ? any Unicode code point except newline ?
unicode_letter      = ? a Unicode code point classified as "Letter" ?
unicode_digit       = ? a Unicode code point classified as "Decimal Digit" ?

identifier_letter   = unicode_letter | "_"
```
### Indentation
Char lexes input following the off-side rule. Any increase in indentation generates an *IDENT* token, and a decrease generates a *DEDENT* as long as the new indentation lines up with a previous indentation level.

Both spaces and tabs are supported; during lexing, tabs are treated as four spaces. Tabs are recommended, but this is not enforced by the compiler.
## Lexical Elements
### Comments
```
line_comment    = ";" { unicode_char } newline
block_comment   = ";;" { unicode_char | newline } ";;"
```
```
; A single-line comment
;; This is
a block
comment ;;
```
### Identifiers
```
identifier  = unicode_letter { identifier_letter | unicode_digit }
```
```
a
b12
another_ident2
αβ
```
### Keywords
### Operators
## Control Structures
### Conditionals
```
"if" expression [ "with" expression ]
    ? code to execute ?

"if" [ expression ] [ "with" expression ]
    "is" expression
        ? code to execute ?
    "is" expression "," expression
        ? code to execute ?
    "is _"
        ? default code to execute ?
```
### Loops
```
```