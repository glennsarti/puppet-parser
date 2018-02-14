# The PN (Pops Notation) format

### Objective

The primary objective for this format is to provide a short but precise
human readable format of the parser AST. This format is needed to write
tests that verifies the parsers function.

Another objective is to provide easy transport of the AST using JSON
or similar protocol that are confined to boolean, number, string,
undef, array, and object. The PN conversion to `Data` was specifically
designed with this objective in mind.

### Format
A PN forms a directed acyclig graph of nodes. There are five types of nodes:

* Literal: A boolean, integer, float, string, or undef
* List: An ordered list of nodes
* Map: An ordered map of string to node associations
* Call: A named list of nodes.

### PN represented as Data

A PN can represent itself as the Puppet `Data` type and will then use
the following conversion rules:

PN Node | Data type
--------|----------
`Literal` | `Boolean`, `Integer`, `Float`, `String`, and `Undef`
`List` | `Array[Data]`
`Map` | `Struct['#', Array[Data]]`
`Call` | `Struct['^', Tuple[String[1], Data, 1]]`

The `Map` is converted into a single element `Hash` with using the key
`'#'` and then an array with an even number of elements where evenly
positioned elements are keys for the value at the next position. The
reason for this is that JSON (and other formats) will not retain the
order of a hash and in Puppet that order is significant.

A `Call` is similar to a `Map` but uses the key `'^'`. The first element
of the associated array is the name of the function and any subsequent
elements are arguments to that function.

### PN represented as String

The native string representation of a PN is similar to Clojure.

PN Node | Sample string representation
--------|----------
`Literal boolean` | `true`, `false`
`Literal integer` | `834`, `-123`, `0`
`Literal float` | `32.28`, `-1.0`, `33.45e18`
`Literal string` | `"plain"`, `"quote \\""`, `"tab \\t"`, `"return \\r"`, `"newline \\n"`, `"control \\u{14}"`
`Literal undef` | `nil`
`List` | `["a" "b" 32 true]`
`Map` | `{:a 2 :b 3 :c true}`
`Call` | `(myFunc 1 2 "b")`

### PN represented as JSON or YAML

When representing PN as JSON or YAML it must first be converted to `Data`. For JSON, this
means that literals are represented verbatim nad lists will be a JSON arrays.

Examples:
A PN `Map` represented as JSON:

    {:a 2 :b 3 :c true} => { "#": ["a", 2, "b", 3, "c", true] }

A PN `Call` represented as JSON:

    (myFunc 1 2 "b") => { "^": [ "myFunc", 1, 2, "b" ] }
