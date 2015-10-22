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
newline           = ? the Unicode code point U+000A ?
unicode_char      = ? any Unicode code point except newline ?
unicode_letter    = ? a Unicode code point classified as "Letter" ?
unicode_digit     = ? a Unicode code point classified as "Decimal Digit" ?
identifier_letter = unicode_letter | "_"
```
### Indentation
Char lexes input following the off-side rule. Any increase in indentation generates an *IDENT* token, and a decrease generates a *DEDENT* as long as the new indentation lines up with a previous indentation level.
```
IDENT  = ? an increase in indentation ?
DEDENT = ? a decrease in indentation ?
```
Both spaces and tabs are supported; during lexing, tabs are treated as four spaces. Tabs are recommended, but this is not enforced by the compiler.
## Lexical Elements
### Comments
```
line_comment  = ";" { unicode_char } newline
block_comment = ";;" { unicode_char | newline } ";;"
```
```
; A single-line comment
;; This is
a block
comment ;;
```
### Identifiers
```
identifier = identifier_letter { identifier_letter | unicode_digit }
```
```
a
b12
another_ident2
αβ
```
### Keywords
```
```
### Operators
```
binary_op     = add_op | multiply_op | boolean_op
boolean_op    = "==" | "!=" | ">" | "<" | ">=" | "<="
add_op        = "+" | "-"
multiply_op   = "*" | "/" | "%"
assignment_op = [ add_op | multiply_op] "="
```
### Expressions
```
expression      = boolean_expr | identifier_expr | binary_expr
boolean_expr    = "true" | "false"
identifier_expr = identifier { "." identifier }
binary_expr     = expression binary_op expression
assignment_expr = identifier_expr assignment_op expression
```
## Statements
### Conditionals
```
if_stmt    = "if" expression [ "with" expression ] newline INDENT statement { statement } DEDENT
if_is_stmt = "if" [ expression ] [ "with" expression ] newline INDENT is_stmt { is_stmt } DEDENT
is_stmt    = "is" expression { "," expression } newline INDENT statement { statement } DEDENT
```
```
if 2 > x
    print("yes")

if 2 > y with y = calc(3)
    print("yes")

if b
    is 1, 2
        print("one or two")
    is _
        print("not 1 or 2")

if with x = calc(7)
    is x > 3
        print("greater than 3")
    is x < 1
        print("less than 1")
    is _
        print("in between")
```
### Loops
```
```
