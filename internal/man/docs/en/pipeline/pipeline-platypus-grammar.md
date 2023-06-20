# Platypus Grammar
---

Below is the syntax definition for the Platypus language used by the Pipeline processor. With the gradual support of different syntaxes, this document will be adjusted and added or deleted to varying degrees.

## Identifier and Keyword {#identifier-and-keyword}

### Identifier {#identifier}

Identifiers are used to identify objects and can be used to represent a variable, function, etc. The identifier contains keywords; the custom identifier cannot be repeated with the keywords of the Pipeline data processor language;

Identifiers can be composed of numbers (`0-9`), letters (`A-Z a-z`), and underscores (`_`), but the first character cannot be a number and is case-sensitive:

- `_abc`
- `abc`
- `abc1`
- `abc_1_`

Use backticks if you need to start with a letter or use characters other than the above in the identifier:

- `` `1abc` ``
- `` `@some-variable` ``
- `` ` 这是一个表情包变量👍` ``

**special agreement**：

We agree to use the identifier `_` to represent the input data of the Pipeline data processor, and this parameter may be implicitly passed to some built-in functions;

To maintain forward compatibility, `_` will be treated as an alias for `message` when used in the current version.

### Keyword {#keyword}

Keywords are words with special meaning, such as `if`, `elif`, `else`, `for`, `in`, `break`, `continue`, `nil`, etc. These words cannot be used as variables or constants or the name of a function etc.

## Comments {#comments}

Use `#` as line comment character, inline comment is not supported

```python
# This is a line comment
a = 1 # this is a comment line

"""
This is a (multiline) string that replaces comments
"""
a = 2

"comments"
a = 3
```

## Built-in Data Types {#built-in-data-types}

In Platypus, the Pipeline data processor language, by default, the type of the value of a variable can change dynamically, but each value has its data type, which can be one of the **basic types** or * *composite type**

When a variable is not assigned a value, its value is nil, which means no value.

### Basic Type {#basic-type}

#### Integer (int) Type {#int-type}

Integer type length is 64bit, signed, currently only supports writing integer literals in decimal format, such as `-1`, `0`, `1`, `+19`

#### Floating Point (float) Type {#float-type}

The length of the floating-point type is 64bit, signed, and currently only supports writing floating-point literals in decimal, such as `-1.00001`, `0.0`, `1.0`, `+19.0`

#### Boolean (bool) Type {#bool-type}

Boolean literals only have `true` and `false`

#### String (str) Type {#str-type}

String literals can be written with double quotes or single quotes, and multi-line strings can be written using triple double quotes or triple single quotes

- `"hello world"`
- `'hello world'`
- Use `"""` to express multi-line strings

   ```python
   """hello
   world"""
   ```

- Use `'''` to express multi-line strings
  
   ```python
   '''
   hello
   the world
   '''
   ```

### Composite Type {#composite-type}

The map type and list type are different from other types. Multiple variables can point to the same map or list object. The memory copy of the list or map is not performed during assignment, but the memory address of the map/list value is referenced.

#### Map Type {#map-type}

The map type is a key-value structure. (Currently) only the string type can be used as a key, and the data type of the value is not limited.

It can read and write elements in the map through index expressions:

```python
a = {
   "1": [1, "2", 3, nil],
   "2": 1.1,
   "abc": nil,
   "def": true
}

# Since a["1"] is a list object, b just refers to the value of a["1"]
b = a["1"]

"""
At this point a["1"][0] == 1.1
"""
b[0] = 1.1
```

#### List Type {#list-type}

The list type can store any number of values of any type in the list.

It can read and write elements in the list through index expressions:

```python
a = [1, "2", 3.0, false, nil, {"a": 1}]

a = a[0] # a == 1
```

## Operator {#operator}

The following operators are currently supported by Platypus, and the higher the value, the higher the priority:

|Priority|Symbol|Associativity|Description|
|-|-|-|-|
| 1 | `=` | right | assignment; named argument; lowest precedence|
| 2 | `||` | left | logical "or" |
| 3 | `&&` | Left | Logical AND |
| 4 | `>=` | left | condition "greater than or equal to" |
| 4 | `>` | left | condition "greater than" |
| 4 | `!=` | left | condition "not equal to" |
| 4 | `==` | left | condition "equal to" |
| 4 | `<=` | left | condition "less than or equal to" |
| 4 | `<` | left | condition "less than" |
| 5 | `+` | left | arithmetic "plus" |
| 5 | `-` | left | arithmetic "minus" |
| 6 | `*` | left | arithmetic "multiply" |
| 6 | `/` | left | arithmetic "division" |
| 6 | `%` | left | arithmetic "remainder"|
| 7 | `[]` | Left | Subscript operator; can use list subscript or map key to get value|
| 7 | `()` | None | Can change operator precedence; function call|

## Expression {#expr}

Platypus uses the symbol comma `,` as the expression separator, such as the parameter transfer used to call the expression and the expression when the map and list are initialized.

In Platypus, expressions can have values, but **statements must not have values**, that is, statements cannot be used as the left and right operands of operators such as `=`, but expressions can.

### Literal Expression {#list-expr}

Literals of various data types can be used as expressions, such as integers `100`, `-1`, `0`, floating-point numbers `1.1`, Boolean values `true`, `false`, etc.

The following two are literal expressions of composite types:

- List literal expression

```txt
[1, true, "1", nil]
```

- Map literal expression

```txt
{
   "a": 1,
   "b": "2",
}
```

### Call Expression {#call-expr}

The following is a function call to get the number of list elements:

```txt
len([1, 3, "5"])
```

### Binary Expression {#binary-expr}

Binary expressions consist of binary operators and left and right operands.

The current version of the assignment expression is a binary expression, which has a return value; but because the assignment expression may cause some problems, this syntax will be deleted in the future, and the **assignment statement** syntax will be added.

```txt
#0
2 out of 5

# 0.4, promote the type of the left operand to a floating-point number during calculation
2 / 5.0

#true
1 + 2 * 3 == 7 && 1 <= 2


# a = (b = 3), a == 3 due to the right associativity of the `=` operator
b == 3;
a = b = 3

# Note: Since the assignment expression syntax is about to be abolished, please replace it with an assignment statement
b = 3
a = b
```

### Index Expression {#index-expr}

Index expressions use the `[]` subscript operator to operate on the elements of a list/map.

The elements of list or map can be valued or modified and elements can be added to the map through index expressions.
For lists, negative numbers can be used for indexing.

Syntax example:

```py
a = [1, 2 ,3, -1.]
b = {"a": [-1], "b": 2}

a[-1] = -2
b["a"][-1] = a[-1]

# result
# a: [1,2,3,-2]
# b: {"a":[-2],"b":2}
```

### Bracket Expression {#bracket-expr}

Bracket expressions can change the precedence of operand operations in binary expressions, but not associativity:

```txt
# 1 + 2 * 3 == 7

(1 + 2) * 3 # == 9
```

## Statement {#stmt}

All expressions in Platypus can be regarded as value statements. When an expression ends with a statement delimiter `;` or `\n`, it will be regarded as a statement. For example, the following script contains four statements:

```go
len("abc")
1
a = 2; a + 2 * 3 % 2
```

### Value Statement (expression statement) {#value-stmt}

When an expression is followed by a statement delimiter, it can be regarded as a value statement. The following are four legal statements:

```txt
# floats as statements
1.;

# function call expression as statement
len("Hello World!"); len({"a": 1})

# identifiers as statements
abc
```

### Assignment Statement {#assignment-stmt}

Syntax example:

```py

key_a = "key-a"

# The identifier a is used as the left operand, assigning a a list literal
a = [1, nil, 3]

# index expression as left operand
a[0] = 0
a[2] = {"key-b": "value-b"}
a[2][key_a] = 123
```

### Select Statement {#select-stmt}

Platypus supports `if/elif/else` syntax:

```txt
if condition {

}
```

```txt
if condition {

} else {

}
```

```txt
if condition_1 {

} elif condition_2 {

} ... elif condition_n {

} else {

}
```

Same as most programming languages, enter the corresponding statement block according to whether the condition of `if/elif` is true, or enter the else branch if it is not true.

The current condition can be any expression, as long as its value is one of the built-in data types, when its value is the default value of the type, the expression value is `flase`:

- When the condition is an `int` type value, it is `0` then the condition is `false`, otherwise it is `true`
- When the condition is a `float` type value, it is `0.0` then the condition is `false`, otherwise it is `true`
- When the condition is a `string` type value, it is an empty string `""` then the condition is `false`, otherwise it is `true`
- when the condition is a `bool` type value, the condition is the current value
- When the condition is a value of type `nil`, the condition is `false`
- When the condition is a `map` type value, its length is 0 then the condition is `false`, otherwise it is `true`
- When the condition is a `list` type value, its length is 0 then the condition is `false`, otherwise it is `true`

### Loop Statement {#loop-stmt}

Platypus supports `for` statement and `for in` statement.

The following are two statements that are only allowed in a loop block:

- `cotinue` statement, do not execute subsequent statements, continue to start the next cycle
- The `break` statement, which ends the loop

When using the `for` statement, it may cause an infinite loop, so it should be used with caution, or use the `for in` statement instead.

```txt
for init-expr; condition; loop-expr {

}
```

```txt
for varb_name in map_value/list_value/string_value {

}
```

Example usage:

- Use `for` to execute 10 loops:

   ```py
   for a = 0; a < 10; a = a + 1 {
    
   }
   ```

- Use `for in` to iterate over all elements of the list:

   ```py
   b = "2"
   for a in ["1", "a" ,"2"] {
     b = b + a
     if b == "21a" {
       break
     }
   }
   # b == "21a"
   ```

- use `for in` to iterate over all keys of the map:

   ```py
   d = 0
   map_a = {"a": 1, "b": 2}
   for x in map_a {
     d = d + map_a[x]
   }
   ```

- use `for in` to iterate over all characters of string:

   ```py
   s = ""
   for c in "abcdef" {
     if s == "abc" {
       break
     } else {
       continue
     }
     s = s + "a"
   }
   # s == "abc"
   ```
