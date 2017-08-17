package parser

import (
  `testing`
  `bytes`
  . "github.com/puppetlabs/go-parser/internal/testutils"
)

func TestEmpty(t *testing.T) {
  expectBlock(t, ``, `(block)`)
}

func TestInvalidUnicode(t *testing.T) {
  expectError(t, "$var = \"\xa0\xa1\"", `invalid unicode character at offset 8`)
  expectError(t, "$var = 23\xa0\xa1", `invalid unicode character at offset 9`)
}

func TestInteger(t *testing.T) {
  expectDump(t, `0`, `0`)
  expectDump(t, `123`, `123`)
  expectDump(t, `+123`, `123`)
  expectDump(t, `0XABC`, `2748`)
  expectDump(t, `0772`, `506`)
  expectError(t, `3g`, `digit expected at line 1:2`)
  expectError(t, `3ö`, `digit expected at line 1:2`)
  expectError(t, `0x3g21`, `hexadecimal digit expected at line 1:4`)
  expectError(t, `078`, `octal digit expected at line 1:3`)
}

func TestNegativeInteger(t *testing.T) {
  expectDump(t, `-123`, `-123`)
}

func TestFloat(t *testing.T) {
  expectDump(t, `0.123`, `0.123`)
  expectDump(t, `123.32`, `123.32`)
  expectDump(t, `+123.32`, `123.32`)
  expectDump(t, `-123.32`, `-123.32`)
  expectDump(t, `12e12`, `1.2e+13`)
  expectDump(t, `12e-12`,`1.2e-11`)
  expectDump(t, `12.23e12`, `1.223e+13`)
  expectDump(t, `12.23e-12`, `1.223e-11`)

  expectError(t, `123.a`, `digit expected at line 1:5`)
  expectError(t, `123.4a`, `digit expected at line 1:6`)

  expectError(t, `123.45ex`, `digit expected at line 1:8`)
  expectError(t, `123.45e3x`, `digit expected at line 1:9`)
}

func TestBoolean(t *testing.T) {
  expectDump(t, `false`, `false`)
  expectDump(t, `true`, `true`)
}

func TestDefault(t *testing.T) {
  expectDump(t, `default`, `(default)`)
}

func TestUndef(t *testing.T) {
  expectDump(t, `undef`, `(undef)`)
}

func TestSingleQuoted(t *testing.T) {
  expectDump(t, `'undef'`, `"undef"`)
  expectDump(t, `'escaped single \''`, `"escaped single '"`)
  expectDump(t, `'unknown escape \k'`, `"unknown escape \\k"`)
}

func TestDoubleQuoted(t *testing.T) {
  expectDump(t,
    `"string\nwith\t\\t,\s\\s, \\r, and \\n\r\n"`,
    `(concat "string\nwith\t\\t, \\s, \\r, and \\n\r\n")`)

  expectDump(t,
    `"unknown \k escape"`,
    `(concat "unknown \\k escape")`)

  expectDump(t,
    `"$var"`,
    `(concat (str ($ var)))`)

  expectDump(t,
    `"hello $var"`,
    `(concat "hello " (str ($ var)))`)

  expectDump(t,
    `"hello ${var}"`,
    `(concat "hello " (str ($ var)))`)

  expectDump(t,
    `"hello ${}"`,
    `(concat "hello " (str (undef)))`)

  expectDump(t,
    `"Before ${{ a => true, b => "hello"}} and after"`,
    `(concat "Before " (str (hash (=> (qn a) true) (=> (qn b) (concat "hello")))) " and after")`)

  expectDump(t, `"x\u{1f452}y"`, `(concat "x👒y")`)

  expectError(t,
    `"$Var"`,
    `malformed interpolation expression at line 1:2`)

  expectError(t,
    Unindent(`
      $x = "y
      notice($x)`),
    "unterminated double quoted string at line 1:6")

  expectError(t,
    Unindent(`
      $x = "y`),
    "unterminated double quoted string at line 1:6")

  expectDump(t, `"x\u2713y"`, `(concat "x✓y")`)
}

func TestRegexp(t *testing.T) {
  expectDump(t,
    `$a = /.*/`,
    `(= ($ a) (regexp ".*"))`)

  expectDump(t, `/pattern\/with\/slash/`, `(regexp "pattern/with/slash")`)
  expectDump(t, `/pattern\/with\\\/slash/`, `(regexp "pattern/with\\\\/slash")`)
  expectDump(t, `/escaped \t/`, `(regexp "escaped \\t")`)

  expectDump(t,
    Unindent(`
      /escaped #rx comment
      continues
      .*/`),
    `(regexp "escaped #rx comment\ncontinues\n.*")`)

  expectError(t,
    `$a = /.*`,
    `unexpected token '/' at line 1:6`)
}

func TestReserved(t *testing.T) {
  expectDump(t,
    `$a = attr`,
    `(= ($ a) (reserved attr))`)

  expectDump(t,
    `$a = private`,
    `(= ($ a) (reserved private))`)
}

func TestHeredoc(t *testing.T) {
  expectHeredoc(t, Unindent(`
      @(END)
      This is
      heredoc text
      END`),
    "This is\nheredoc text\n")

  expectError(t, Unindent(`
      @(END)
      This is
      heredoc text`),
    "unterminated heredoc at line 1:1")

  expectDump(t,
    Unindent(`
      { a => @(ONE), b => @(TWO) }
      The first
      heredoc text
      -ONE
      The second
      heredoc text
      -TWO`),
    `(hash (=> (qn a) (heredoc {:text "The first\nheredoc text"})) (=> (qn b) (heredoc {:text "The second\nheredoc text"})))`)

  expectDump(t,
    Unindent(`
      ['first', @(SECOND), 'third', @(FOURTH), 'fifth',
        This is the text of the
        second entry
        |-SECOND
        And here is the text of the
        fourth entry
        |-FOURTH
        'sixth']`),
    `(array "first" (heredoc {:text "This is the text of the\nsecond entry"}) "third" (heredoc {:text "And here is the text of the\nfourth entry"}) "fifth" "sixth")`)

  expectError(t,
    Unindent(`
      @(END
      /t)
      This\nis\nheredoc\ntext
      -END`),
    `unterminated @( at line 1:1`)

  expectError(t,
    Unindent(`
      @(END)
      This\nis\nheredoc\ntext

      `),
    `unterminated heredoc at line 1:1`)

  expectError(t,
    Unindent(`
      @(END)`),
    `unterminated heredoc at line 1:1`)
}

func TestHeredocSyntax(t *testing.T) {
  expectDump(t, Unindent(`
      @(END:syntax)
      This is
      heredoc text
      END`),
    `(heredoc {:syntax syntax :text "This is\nheredoc text\n"})`)

  expectError(t, Unindent(`
      @(END:json:yaml)
      This is
      heredoc text`),
    `more than one syntax declaration in heredoc at line 1:11`)
}

func TestHeredocFlags(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @(END/t)
      This\tis\t
      heredoc text
      -END`),
    "This\tis\t\nheredoc text")

  expectHeredoc(t,
    Unindent(`
      @(END/s)
      This\sis\sheredoc\stext
      -END`),
    `This is heredoc text`)

  expectHeredoc(t,
    Unindent(`
      @(END/r)
      This\ris\rheredoc\rtext
      -END`),
    "This\ris\rheredoc\rtext")

  expectHeredoc(t,
    Unindent(`
      @(END/n)
      This\nis\nheredoc\ntext
      -END`),
    "This\nis\nheredoc\ntext")

  expectHeredoc(t,
    Unindent(`
      @(END:syntax/n)
      This\nis\nheredoc\ntext
      -END`),
    "This\nis\nheredoc\ntext", `syntax`)

  expectError(t,
    Unindent(`
      @(END/k)
      This\nis\nheredoc\ntext
      -END`),
    `illegal heredoc escape 'k' at line 1:7`)

  expectError(t,
    Unindent(`
      @(END/t/s)
      This\nis\nheredoc\ntext
      -END`),
    `more than one declaration of escape flags in heredoc at line 1:8`)
}

func TestHeredocStripNL(t *testing.T) {
  expectHeredoc(t,
    "@(END)\r\nThis is\r\nheredoc text\r\n-END",
    "This is\r\nheredoc text")
}

func TestHeredocMargin(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @(END/t)
        This\tis
        heredoc text
        | END
      `),
    "This\tis\nheredoc text\n")

  // Lines that have less margin than what's stripped are not stripped
  expectHeredoc(t,
    Unindent(`
      @(END/t)
        This\tis
       heredoc text
        | END
      `),
    "This\tis\n heredoc text\n")
}

func TestHeredocMarginAndNewlineTrim(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @(END/t)
        This\tis
        heredoc text
        |- END`),
    "This\tis\nheredoc text")
}

func TestHeredocInterpolate(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @("END")
        This is
        heredoc $text
        |- END`),
    `(heredoc {:text (concat "This is\nheredoc " (str ($ text)))})`)

  expectHeredoc(t,
    Unindent(`
      @("END")
        This is
        heredoc $a \$b
        |- END`),
    `(heredoc {:text (concat "This is\nheredoc " (str ($ a)) " \\" (str ($ b)))})`)

  expectHeredoc(t,
    Unindent(`
      @("END"/$)
        This is
        heredoc $a \$b
        |- END`),
    `(heredoc {:text (concat "This is\nheredoc " (str ($ a)) " $b")})`)

  expectHeredoc(t,
    Unindent(`
      @(END)
        This is
        heredoc $text
        |- END`),
    Unindent(`
      This is
      heredoc $text`))

  expectError(t,
    Unindent(`
      @("END""MORE")
        This is
        heredoc $text
        |- END`),
    `more than one tag declaration in heredoc at line 1:8`)

  expectError(t,
    Unindent(`
      @("END
      ")
        This is
        heredoc $text
        |- END`),
    `unterminated @( at line 1:1`)

  expectError(t,
    Unindent(`
      @("")
        This is
        heredoc $text
        |-`),
    `empty heredoc tag at line 1:1`)

  expectError(t,
    Unindent(`
      @()
        This is
        heredoc $text
        |-`),
    `empty heredoc tag at line 1:1`)
}

func TestHeredocNewlineEscape(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @(END/L)
        Do not break \
        this line
        |- END`),
    Unindent(`
      Do not break this line`))

  expectHeredoc(t,
    Unindent(`
      @(END/L)
        Do not break \
        this line\
        |- END`),
    Unindent(`
      Do not break this line\`))

  expectHeredoc(t,
    Unindent(`
      @(END/t)
        Do break \
        this line
        |- END`),
    Unindent(`
      Do break \
      this line`))

  expectHeredoc(t,
    Unindent(`
      @(END/u)
        A checkmark \u2713 symbol
        |- END`),
    Unindent(`
      A checkmark ✓ symbol`))
}

func TestHeredocUnicodeEscape(t *testing.T) {
  expectHeredoc(t,
    Unindent(`
      @(END/u)
        A hat \u{1f452} symbol
        |- END`),
    Unindent(`
      A hat 👒 symbol`))

  expectHeredoc(t,
    Unindent(`
      @(END/u)
        A checkmark \u2713 symbol
        |- END`),
    Unindent(`
      A checkmark ✓ symbol`))

  expectError(t,
    Unindent(`
      @(END/u)
        A hat \u{1f452 symbol
        |- END`),
    `malformed unicode escape sequence at line 2:9`)

  expectError(t,
    Unindent(`
      @(END/u)
        A hat \u{1f45234} symbol
        |- END`),
    `malformed unicode escape sequence at line 2:9`)

  expectError(t,
    Unindent(`
      @(END/u)
        A hat \u{1} symbol
        |- END`),
    `malformed unicode escape sequence at line 2:9`)

  expectError(t,
    Unindent(`
      @(END/u)
        A checkmark \u271 symbol
        |- END`),
    `malformed unicode escape sequence at line 2:15`)

  expectError(t,
    Unindent(`
      @(END/u)
        A checkmark \ux271 symbol
        |- END`),
    `malformed unicode escape sequence at line 2:15`)
}

func TestMLCommentAfterHeredocTag(t *testing.T) {
  expectHeredoc(t, Unindent(`
      @(END) /* comment after tag */
      This is
      heredoc text
      END`),
    "This is\nheredoc text\n")
}

func TestCommentAfterHeredocTag(t *testing.T) {
  expectHeredoc(t, Unindent(`
      @(END) # comment after tag
      This is
      heredoc text
      END`),
    "This is\nheredoc text\n")
}

func TestVariable(t *testing.T) {
  expectDump(t,
    `$var`,
    `($ var)`)

  expectDump(t,
    `$var::b`,
    `($ var::b)`)

  expectDump(t,
    `$::var::b`,
    `($ ::var::b)`)

  expectDump(t,
    `$::var::_b`,
    `($ ::var::_b)`)

  expectDump(t,
    `$2`,
    `($ 2)`)

  expectDump(t,
    `$`,
    `($ )`)

  expectError(t,
    `$var:b`,
    `unexpected token ':' at line 1:5`)

  expectError(t,
    `$Var`,
    `invalid variable name at line 1:2`)

  expectError(t,
    `$:var::b`,
    `invalid variable name at line 1:1`)

  expectError(t,
    `$::var::B`,
    `invalid variable name at line 1:1`)

  expectError(t,
    `$::var::_b::c`,
    `invalid variable name at line 1:1`)

  expectError(t,
    `$::_var::b`,
    `unexpected token '_' at line 1:4`)
}

func TestArray(t *testing.T) {
  expectDump(t,
    `[1,2,3]`,
    `(array 1 2 3)`)

  expectDump(t,
    `[1,2,3,]`,
    `(array 1 2 3)`)

  expectError(t,
    `[1,2 3]`,
    `expected one of ',' or ']', got 'integer literal' at line 1:6`)

  expectError(t,
    `[1,2,3`,
    `expected one of ',' or ']', got 'EOF' at line 1:7`)
}

func TestHash(t *testing.T) {
  expectDump(t,
    `{ a => true, b => false, c => undef, d => 12, e => 23.5, c => 'hello' }`,
    `(hash (=> (qn a) true) (=> (qn b) false) (=> (qn c) (undef)) (=> (qn d) 12) (=> (qn e) 23.5) (=> (qn c) "hello"))`)

  expectDump(t,
    `{a => 1, b => 2,}`,
    `(hash (=> (qn a) 1) (=> (qn b) 2))`)

  expectError(t,
    `{a => 1, b, 2}`,
    `expected '=>' to follow hash key at line 1:12`)

  expectError(t,
    `{a => 1 b => 2}`,
    `expected one of ',' or '}', got 'identifier' at line 1:9`)

  expectError(t,
    `{a => 1, b => 2`,
    `expected one of ',' or '}', got 'EOF' at line 1:16`)
}

func TestBlock(t *testing.T) {
  expectBlock(t,
    Unindent(`
      $t = 'the'
      $r = 'revealed'
      $map = {'ipl' => 'meaning', 42.0 => 'life'}
      "$t ${map['ipl']} of ${map[42.0]}${[3, " is not ${r}"][1]} here"`),
    `(block `+
        `(= ($ t) "the") `+
        `(= ($ r) "revealed") `+
        `(= ($ map) (hash (=> "ipl" "meaning") (=> 42 "life"))) `+
        `(concat (str ($ t)) " " (str ([] ($ map) "ipl")) " of " (str ([] ($ map) 42)) (str ([] (array 3 (concat " is not " (str ($ r)))) 1)) " here"))`)

  expectBlock(t,
    Unindent(`
      $t = 'the';
      $r = 'revealed';
      $map = {'ipl' => 'meaning', 42.0 => 'life'};
      "$t ${map['ipl']} of ${map[42.0]}${[3, " is not ${r}"][1]} here"`),
    `(block `+
        `(= ($ t) "the") `+
        `(= ($ r) "revealed") `+
        `(= ($ map) (hash (=> "ipl" "meaning") (=> 42 "life"))) `+
        `(concat (str ($ t)) " " (str ([] ($ map) "ipl")) " of " (str ([] ($ map) 42)) (str ([] (array 3 (concat " is not " (str ($ r)))) 1)) " here"))`)
}

func TestFunctionDefintion(t *testing.T) {
  expectDump(t,
    Unindent(`
      function myFunc(Integer[0,3] $first, $untyped, String $nxt = 'hello') >> Float {
         23.8
      }`),
    `(function {`+
      `:name myFunc `+
      `:params [`+
        `{:name first :type ([] (qr Integer) 0 3)} `+
        `{:name untyped} `+
        `{:name nxt :type (qr String) :value "hello"}] `+
      `:returns (qr Float) `+
      `:body [23.8]})`)

  expectDump(t,
    Unindent(`
      function myFunc(Integer *$numbers) >> Integer {
         $numbers.size
      }`),
    `(function {`+
        `:name myFunc `+
        `:params [`+
          `{:name numbers :type (qr Integer) :splat true}] `+
        `:returns (qr Integer) `+
        `:body [`+
          `(call_method {:functor (. ($ numbers) (qn size)) :args []})]})`)

  expectError(t,
    Unindent(`
      function myFunc(Integer *numbers) >> Integer {
         numbers.size
      }`),
    `expected variable declaration at line 1:33`)

  expectError(t,
    Unindent(`
      function myFunc(Integer *$numbers) >> $var {
         numbers.size
      }`),
    `expected type name at line 1:43`)

  expectError(t,
    Unindent(`
      function 'myFunc'() {
         true
      }`),
    `expected a name to follow keyword 'function' at line 1:10`)

  expectError(t,
    Unindent(`
      function myFunc() true`),
    `expected token '{', got 'boolean literal' at line 1:19`)

  expectError(t,
    Unindent(`
      function myFunc() >> Boolean true`),
    `expected token '{', got 'boolean literal' at line 1:30`)
}

func TestNodeDefinition(t *testing.T) {
  expectDump(t,
    Unindent(`
      node default {
      }`),
    `(node {:matches [(default)] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/ {
      }`),
    `(node {:matches [(regexp "[a-f].*")] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/, "example.com" {
      }`),
    `(node {:matches [(regexp "[a-f].*") (concat "example.com")] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/, example.com {
      }`),
    `(node {:matches [(regexp "[a-f].*") "example.com"] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/, 192.168.0.1, 34, "$x.$y" {
      }`),
    `(node {:matches [(regexp "[a-f].*") "192.168.0.1" "34" (concat (str ($ x)) "." (str ($ y)))] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/, 192.168.0.1, 34, "$x.$y", {
      }`),
    `(node {:matches [(regexp "[a-f].*") "192.168.0.1" "34" (concat (str ($ x)) "." (str ($ y)))] :body []})`)

  expectDump(t,
    Unindent(`
      node /[a-f].*/ inherits 192.168.0.1 {
      }`),
    `(node {:matches [(regexp "[a-f].*")] :parent "192.168.0.1" :body []})`)

  expectDump(t,
    Unindent(`
      node default {
        notify { x: message => 'node default' }
      }`),
    `(node {:matches [(default)] :body [(resource {:type (qn notify) :bodies [{:title (qn x) :ops [(=> (message) "node default")]}]})]})`)

  expectError(t,
    Unindent(`
      node [hosta.com, hostb.com] {
      }`),
    Unindent(`hostname expected at line 1:7`))

  expectError(t,
    Unindent(`
      node example.* {
      }`),
    Unindent(`expected name or number to follow '.' at line 1:15`))
}

func TestSiteDefinition(t *testing.T) {
  expectDump(t,
    Unindent(`
      site {
      }`),
    `(site (block))`)

  expectDump(t,
    Unindent(`
      site {
        notify { x: message => 'node default' }
      }`),
    `(site (block (resource {:type (qn notify) :bodies [{:title (qn x) :ops [(=> (message) "node default")]}]})))`)
}

func TestTypeDefinition(t *testing.T) {
  expectDump(t,
    Unindent(`
      type MyType {
        # What statements that can be included here is not yet speced
      }`),
    `(type-definition (MyType) () (block))`)

  expectDump(t,
    Unindent(`
      type MyType inherits OtherType {
        # What statements that can be included here is not yet speced
      }`),
    `(type-definition (MyType) (OtherType) (block))`)

  expectError(t,
    Unindent(`
      type MyType inherits OtherType [{
        # What statements that can be included here is not yet speced
      }]`),
    `expected token '{', got '[' at line 1:32`)

  expectError(t,
    Unindent(`
      type MyType inherits $other {
        # What statements that can be included here is not yet speced
      }`),
    `expected type name to follow 'inherits' at line 1:28`)

  expectError(t,
    Unindent(`
      type MyType[a,b] {
        # What statements that can be included here is not yet speced
      }`),
    `expected type name to follow 'type' at line 1:19`)

  expectError(t,
    Unindent(`
      type MyType << {
        # What statements that can be included here is not yet speced
      }`),
    `unexpected token '<<' at line 1:15`)
}

func TestTypeAlias(t *testing.T) {
  expectDump(t,
    Unindent(`
      type MyType = Object[{
        attributes => {
        name => String,
        number => Integer
        }
      }]`),
    `(type-alias (MyType) ([] (qr Object) (hash (=> (qn attributes) (hash (=> (qn name) (qr String)) (=> (qn number) (qr Integer)))))))`)

  expectError(t,
    Unindent(`
      type MyType[a, b] = Object[{
        attributes => {
        name => String,
        number => Integer
        }
      }]`),
    `expected type name to follow 'type' at line 1:20`)

  expectError(t,
    `type Mod::myType[a, b] = Object[{}]`,
    `invalid type name at line 1:6`)
}

func TestClass(t *testing.T) {
  expectDump(t,
    Unindent(`
      class myclass {
      }`),
    `(class {:name myclass :body []})`)

  expectDump(t,
    Unindent(`
      class myclass {
        class inner {
        }
      }`),
    `(class {:name myclass :body [(class {:name myclass::inner :body []})]})`)

  expectDump(t,
    Unindent(`
      class ::myclass {
        class inner {
        }
      }`),
    `(class {:name myclass :body [(class {:name myclass::inner :body []})]})`)

  expectDump(t,
    Unindent(`
      class ::myclass {
        class ::inner {
        }
      }`),
    `(class {:name myclass :body [(class {:name myclass::inner :body []})]})`)

  expectDump(t,
    Unindent(`
      class myclass inherits other {
      }`),
    `(class {:name myclass :parent other :body []})`)

  expectDump(t,
    Unindent(`
      class myclass inherits default {
      }`),
    `(class {:name myclass :parent default :body []})`)

  expectDump(t,
    Unindent(`
      class myclass($a, $b = 2) {
      }`),
    `(class {:name myclass :params [{:name a} {:name b :value 2}] :body []})`)

  expectDump(t,
    Unindent(`
      class myclass($a, $b = 2) inherits other {
      }`),
    `(class {:name myclass :parent other :params [{:name a} {:name b :value 2}] :body []})`)

  expectError(t,
    Unindent(`
      class 'myclass' {
      }`),
    `a quoted string is not valid as a name at this location at line 1:7`)

  expectError(t,
    Unindent(`
      class class {
      }`),
    `'class' keyword not allowed at this location at line 1:7`)

  expectError(t,
    Unindent(`
      class [a, b] {
      }`),
    `expected name of class at line 1:7`)
}

func TestDefinition(t *testing.T) {
  expectDump(t,
    Unindent(`
      define apache::vhost (
        Integer $port,
        String[1] $docroot,
        String[1] $servername = $title,
        String $vhost_name = '*',
      ) {
        include apache # contains package['httpd'] and service['httpd']
        include apache::params # contains common config settings

        $vhost_dir = $apache::params::vhost_dir

        # the template used below can access all of the parameters and variable from above.
        file { "${vhost_dir}/${servername}.conf":
          ensure  => file,
          owner   => 'www',
          group   => 'www',
          mode    => '0644',
          content => template('apache/vhost-default.conf.erb'),
          require => Package['httpd'],
          notify  => Service['httpd'],
        }
      }`),
    `(define {`+
      `:name apache::vhost `+
      `:params [`+
        `{:name port :type (qr Integer)} `+
        `{:name docroot :type ([] (qr String) 1)} `+
        `{:name servername :type ([] (qr String) 1) :value ($ title)} `+
        `{:name vhost_name :type (qr String) :value "*"}] `+
      `:body [`+
        `(invoke {:functor (qn include) :args [(qn apache)]}) `+
        `(invoke {:functor (qn include) :args [(qn apache::params)]}) `+
        `(= ($ vhost_dir) ($ apache::params::vhost_dir)) `+
        `(resource {`+
          `:type (qn file) `+
          `:bodies [{`+
            `:title (concat (str ($ vhost_dir)) "/" (str ($ servername)) ".conf") `+
            `:ops [`+
              `(=> (ensure) (qn file)) `+
              `(=> (owner) "www") `+
              `(=> (group) "www") `+
              `(=> (mode) "0644") `+
              `(=> (content) (call {:functor (qn template) :args ["apache/vhost-default.conf.erb"]})) `+
              `(=> (require) ([] (qr Package) "httpd")) `+
              `(=> (notify) ([] (qr Service) "httpd"))]}]})]})`)
}

func TestApplication(t *testing.T) {
  expectDump(t,
    Unindent(`
      MyCap produces Cap {
        attr => $value
      }`),
    `(produces (qr MyCap) (Cap (=> (attr) ($ value))))`)
}

func TestCapabilityMappping(t *testing.T) {
  expectDump(t,
    Unindent(`
      application lamp(
        String $db_user,
        String $db_password,
        String $docroot = '/var/www/html',
      ){
        lamp::web { $name:
          docroot => $docroot,
          export => Http["lamp-${name}"],
        }

        lamp::app { $name:
          consume => Sql["lamp-${name}"],
          export => Http["lamp-${name}"],
        }

        lamp::db { $name:
          db_user     => $db_user,
          db_name     => $db_name,
          export      => Sql["lamp-${name}"],
        }
      }`),

    `(application {`+
      `:name lamp `+
      `:params [`+
        `{:name db_user :type (qr String)} `+
        `{:name db_password :type (qr String)} `+
        `{:name docroot :type (qr String) :value "/var/www/html"}] `+
      `:body [`+
        `(resource {`+
					`:type (qn lamp::web) `+
					`:bodies [{`+
            `:title ($ name) `+
            `:ops [(=> (docroot) ($ docroot)) (=> (export) ([] (qr Http) (concat "lamp-" (str ($ name)))))]}]}) `+
        `(resource {`+
					`:type (qn lamp::app) `+
					`:bodies [{`+
            `:title ($ name) `+
					  `:ops [(=> (consume) ([] (qr Sql) (concat "lamp-" (str ($ name))))) (=> (export) ([] (qr Http) (concat "lamp-" (str ($ name)))))]}]}) `+
        `(resource {`+
					`:type (qn lamp::db) `+
					`:bodies [{`+
            `:title ($ name) `+
					  `:ops [(=> (db_user) ($ db_user)) (=> (db_name) ($ db_name)) (=> (export) ([] (qr Sql) (concat "lamp-" (str ($ name)))))]}]})]})`)
}

func TestCallNamed(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = wrap(myFunc(3, 'vx', 'd"x') |Integer $r| >> Integer { $r + 2 })`),

      `(= ($ x) (call {`+
        `:functor (qn wrap) `+
        `:args [(call {`+
          `:functor (qn myFunc) `+
          `:args [3 "vx" "d\"x"] `+
          `:block {`+
            `:params [{:name r :type (qr Integer)}] `+
            `:returns (qr Integer) `+
            `:body [(+ ($ r) 2)]}})]}))`)

  expectDump(t,
    `notice hello()`, `(invoke {:functor (qn notice) :args [(call {:functor (qn hello) :args []})]})`)

  expectDump(t,
    `notice hello(), 'world'`, `(invoke {:functor (qn notice) :args [(call {:functor (qn hello) :args []}) "world"]})`)

  expectBlock(t,
    Unindent(`
      $x = $y.myFunc
      callIt(*$x)
      (2 + 3).with() |$x| { notice $x }`),
    `(block `+
      `(= ($ x) (call_method {:functor (. ($ y) (qn myFunc)) :args []})) `+
      `(invoke {:functor (qn callIt) :args [(unfold ($ x))]}) `+
      `(call_method {`+
        `:functor (. (() (+ 2 3)) (qn with)) `+
        `:args [] `+
        `:block {`+
          `:params [{:name x}] `+
          `:body [(invoke {`+
            `:functor (qn notice) `+
            `:args [($ x)]})]}}))`)

  expectError(t,
    Unindent(`
      $x = myFunc(3`),
    `expected one of ',' or ')', got 'EOF' at line 1:14`)

  expectError(t,
    Unindent(`
      $x = myFunc() || $r + 2 }`),
    `expected token '{', got 'variable' at line 1:18`)

}

func TestCallNamedNoArgs(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = wrap(myFunc |Integer $r| >> Integer { $r + 2 })`),
    `(= ($ x) (call {`+
        `:functor (qn wrap) `+
        `:args [(call {`+
          `:functor (qn myFunc) `+
          `:args [] `+
          `:block {`+
            `:params [{:name r :type (qr Integer)}] `+
            `:returns (qr Integer) `+
            `:body [(+ ($ r) 2)]}})]}))`)
}

func TestCallMethod(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = $y.max(23)`),
    `(= ($ x) (call_method {:functor (. ($ y) (qn max)) :args [23]}))`)
}

func TestCallMethodArgsLambda(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = $y.max(23) |$x| { $x }`),
    `(= ($ x) (call_method {:functor (. ($ y) (qn max)) :args [23] :block {:params [{:name x}] :body [($ x)]}}))`)
}

func TestCallMethodNoArgs(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = $y.max`),
    `(= ($ x) (call_method {:functor (. ($ y) (qn max)) :args []}))`)
}

func TestCallMethodNoArgsLambda(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = $y.max |$x| { $x }`),
    `(= ($ x) (call_method {:functor (. ($ y) (qn max)) :args [] :block {:params [{:name x}] :body [($ x)]}}))`)
}

func TestCallType(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = type(3)`),
    `(= ($ x) (call {:functor (qn type) :args [3]}))`)
}

func TestCallTypeMethod(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = $x.type(3)`),
    `(= ($ x) (call_method {:functor (. ($ x) (qn type)) :args [3]}))`)
}

func TestLineComment(t *testing.T) {
  expectBlock(t,
    Unindent(`
      $x = 'y'
      # The above is a variable assignment
      # and here is a notice of the assigned
      # value
      #
      notice($y)`),
    `(block (= ($ x) "y") (invoke {:functor (qn notice) :args [($ y)]}))`)
}

func TestIdentifiers(t *testing.T) {
  expectDump(t,
    `name`,
    `(qn name)`)

  expectDump(t,
    `Name`,
    `(qr Name)`)

  expectDump(t,
    `Ab::Bc`,
    `(qr Ab::Bc)`)

  expectDump(t,
    `$x = ::assertType(::TheType, $y)`,
    `(= ($ x) (call {:functor (qn ::assertType) :args [(qr ::TheType) ($ y)]}))`)

  expectError(t,
    `abc:cde`,
    `unexpected token ':' at line 1:4`)

  expectError(t,
    `Ab::bc`,
    `invalid type name at line 1:1`)

  expectError(t,
    `$x = ::3m`,
    `:: not followed by name segment at line 1:6`)
}

func TestRestOfLineComment(t *testing.T) {
  expectBlock(t,
    Unindent(`
      $x = 'y' # A variable assignment
      notice($y)`),
    `(block (= ($ x) "y") (invoke {:functor (qn notice) :args [($ y)]}))`)

  expectBlock(t,
    Unindent(`
      # [*version*]
      #   The package version to install, used to set the package name.
      #   Defaults to undefined`),
    `(block)`)
}

func TestMultilineComment(t *testing.T) {
  expectBlock(t,
    Unindent(`
      $x = 'y'
      /* The above is a variable assignment
         and here is a notice of the assigned
         value
      */
      notice($y)`),
    `(block (= ($ x) "y") (invoke {:functor (qn notice) :args [($ y)]}))`)
}

func TestSingleQuote(t *testing.T) {
  expectDump(t, `$x = 'a string'`, `(= ($ x) "a string")`)

  expectDump(t, `$x = 'a \'string\' with \\'`, `(= ($ x) "a 'string' with \\")`)

  expectError(t,
    Unindent(`
      $x = 'y
      notice($x)`),
    "unterminated single quoted string at line 1:6")
}

func TestUnterminatedQuoteEscapedEnd(t *testing.T) {
  expectError(t,
    Unindent(`
      $x = 'y\`),
    "unterminated single quoted string at line 1:6")
}

func TestStrayTilde(t *testing.T) {
  expectError(t,
    Unindent(`
      $x ~ 'y'
      notice($x)`),
    "unexpected token '~' at line 1:4")
}

func TestUnknownToken(t *testing.T) {
  expectError(t,
    Unindent(`
      $x ^ 'y'
      notice($x)`),
    "unexpected token '^' at line 1:4")
}

func TestUnterminatedComment(t *testing.T) {
  expectError(t,
    Unindent(`
      $x = 'y'
      /* The above is a variable assignment
      notice($y)`),
    "unterminated /* */ comment at line 2:1")
}

func TestIf(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = if $y {
        true
      } else {
        false
      }`),
    `(= ($ x) (if {:test ($ y) :then [true] :else [false]}))`)

  expectDump(t,
    Unindent(`
      $x = if $y > 2 {
      } else {
        false
      }`),
    `(= ($ x) (if {:test (> ($ y) 2) :then [] :else [false]}))`)

  expectDump(t,
    Unindent(`
      $x = if $y != 34 {
        true
      } else {
      }`),
    `(= ($ x) (if {:test (!= ($ y) 34) :then [true] :else []}))`)

  expectDump(t,
    Unindent(`
      $x = if $y {
        1
      } elsif $z {
        2
      } else {
        3
      }`),
    `(= ($ x) (if {:test ($ y) :then [1] :else {:test ($ z) :then [2] :else [3]}}))`)

  expectDump(t,
    Unindent(`
      $x = if $y == if $x {
        true
      } { false }`),
    `(= ($ x) (if {:test (== ($ y) (if {:test ($ x) :then [true]})) :then [false]}))`)

  expectError(t,
    `$x = else { 3 }`,
    `unexpected token 'else' at line 1:6`)
}

func TestUnless(t *testing.T) {
  expectDump(t,
    Unindent(`
      $x = unless $y {
        true
      } else {
        false
      }`),
    `(= ($ x) (unless {:test ($ y) :then [true] :else [false]}))`)

  expectDump(t,
    Unindent(`
      $x = unless $y {
      } else {
        false
      }`),
    `(= ($ x) (unless {:test ($ y) :then [] :else [false]}))`)

  expectDump(t,
    Unindent(`
      $x = unless $y {
        true
      } else {
      }`),
    `(= ($ x) (unless {:test ($ y) :then [true] :else []}))`)

  expectDump(t,
    Unindent(`
      $x = if $y == unless $x {
        true
      } { false }`),
    `(= ($ x) (if {:test (== ($ y) (unless {:test ($ x) :then [true]})) :then [false]}))`)

  expectError(t,
    Unindent(`
      $x = unless $y {
        1
      } elsif $z {
        2
      } else {
        3
      }`),
    `elsif not supported in unless expression at line 3:8`)
}

func TestSelector(t *testing.T) {
  expectDump(t,
    `$rootgroup = $facts['os']['family'] ? 'Solaris' => 'wheel'`,
    `(= ($ rootgroup) (? ([] ([] ($ facts) "os") "family") [(=> "Solaris" "wheel")]))`)

  expectDump(t,
    Unindent(`
      $rootgroup = $facts['os']['family'] ? {
        'Solaris'          => 'wheel',
        /(Darwin|FreeBSD)/ => 'wheel',
        default            => 'root'
      }`),
    `(= ($ rootgroup) (? ([] ([] ($ facts) "os") "family") [(=> "Solaris" "wheel") (=> (regexp "(Darwin|FreeBSD)") "wheel") (=> (default) "root")]))`)

  expectDump(t,
    Unindent(`
      $rootgroup = $facts['os']['family'] ? {
        'Solaris'          => 'wheel',
        /(Darwin|FreeBSD)/ => 'wheel',
        default            => 'root',
      }`),
    `(= ($ rootgroup) (? ([] ([] ($ facts) "os") "family") [(=> "Solaris" "wheel") (=> (regexp "(Darwin|FreeBSD)") "wheel") (=> (default) "root")]))`)
}

func TestCase(t *testing.T) {
  expectDump(t,
    Unindent(`
    case $facts['os']['name'] {
      'Solaris':           { include role::solaris } # Apply the solaris class
      'RedHat', 'CentOS':  { include role::redhat  } # Apply the redhat class
      /^(Debian|Ubuntu)$/: { include role::debian  } # Apply the debian class
      default:             { include role::generic } # Apply the generic class
    }`),
    `(case `+
        `{:when ["Solaris"] :then (block (invoke {:functor (qn include) :args [(qn role::solaris)]}))} `+
        `{:when ["RedHat" "CentOS"] :then (block (invoke {:functor (qn include) :args [(qn role::redhat)]}))} `+
        `{:when [(regexp "^(Debian|Ubuntu)$")] :then (block (invoke {:functor (qn include) :args [(qn role::debian)]}))} `+
        `{:when [(default)] :then (block (invoke {:functor (qn include) :args [(qn role::generic)]}))})`)
}

func TestResource(t *testing.T) {
  expectDump(t,
    Unindent(`
      file { '/tmp/foo':
        mode => '0640',
        ensure => present
      }`),
    `(resource {`+
        `:type (qn file) `+
        `:bodies [{:title "/tmp/foo" :ops [(=> (mode) "0640") (=> (ensure) (qn present))]}]})`)

  expectDump(t,
    Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership
      }`),
    `(resource {`+
        `:type (qn file) `+
        `:bodies [{:title "/tmp/foo" :ops [(=> (ensure) (qn file)) (=> (*) ($ file_ownership))]}]})`)

  expectDump(t,
    Unindent(`
      @file { '/tmp/foo':
        mode => '0640',
        ensure => present
      }`),
    `(resource {`+
        `:form virtual `+
        `:type (qn file) `+
        `:bodies [{:title "/tmp/foo" :ops [(=> (mode) "0640") (=> (ensure) (qn present))]}]})`)

  expectDump(t,
    Unindent(`
      @@file { '/tmp/foo':
        mode => '0640',
        ensure => present
      }`),
    `(resource {`+
        `:form exported `+
        `:type (qn file) `+
        `:bodies [{:title "/tmp/foo" :ops [(=> (mode) "0640") (=> (ensure) (qn present))]}]})`)

  expectDump(t,
    Unindent(`
      class { some_title: }`),
    `(resource {:type (qn class) :bodies [{:title (qn some_title) :ops []}]})`)

  expectDump(t,
    Unindent(`
      file { '/tmp/foo': }`),
    `(resource {`+
        `:type (qn file) `+
        `:bodies [{:title "/tmp/foo" :ops []}]})`)

  expectDump(t,
    Unindent(`
      package { 'openssh-server':
        ensure => present,
      } -> # and then:
      file { '/etc/ssh/sshd_config':
        ensure => file,
        mode   => '0600',
        source => 'puppet:///modules/sshd/sshd_config',
      } ~> # and then:
      service { 'sshd':
        ensure => running,
        enable => true,
      }`),
    `(~> (-> `+
      `(resource {`+
        `:type (qn package) `+
        `:bodies [{`+
          `:title "openssh-server" `+
          `:ops [(=> (ensure) (qn present))]}]}) `+
      `(resource {`+
        `:type (qn file) `+
        `:bodies [{`+
          `:title "/etc/ssh/sshd_config" `+
          `:ops [(=> (ensure) (qn file)) (=> (mode) "0600") (=> (source) "puppet:///modules/sshd/sshd_config")]}]})) `+
      `(resource {`+
        `:type (qn service) `+
        `:bodies [{`+
          `:title "sshd" `+
          `:ops [(=> (ensure) (qn running)) (=> (enable) true)]}]}))`)

  expectDump(t,
    Unindent(`
      package { 'openssh-server':
        ensure => present,
      } <- # and then:
      file { '/etc/ssh/sshd_config':
        ensure => file,
        mode   => '0600',
        source => 'puppet:///modules/sshd/sshd_config',
      } <~ # and then:
      service { 'sshd':
        ensure => running,
        enable => true,
      }`),
    `(<~ (<- `+
      `(resource {`+
        `:type (qn package) `+
        `:bodies [{`+
          `:title "openssh-server" `+
          `:ops [(=> (ensure) (qn present))]}]}) `+
      `(resource {`+
        `:type (qn file) `+
        `:bodies [{`+
          `:title "/etc/ssh/sshd_config" `+
          `:ops [(=> (ensure) (qn file)) (=> (mode) "0600") (=> (source) "puppet:///modules/sshd/sshd_config")]}]})) `+
      `(resource {`+
        `:type (qn service) `+
        `:bodies [{`+
          `:title "sshd" `+
          `:ops [(=> (ensure) (qn running)) (=> (enable) true)]}]}))`)

  expectError(t,
    Unindent(`
      file { '/tmp/foo':
        mode => '0640',
        ensure => present
      `),
    `expected token '}', got 'EOF' at line 4:1`)

  expectError(t,
    Unindent(`
      file { '/tmp/foo':
        mode, '0640',
        ensure, present
      }`),
    `invalid attribute operation at line 2:8`)

  expectError(t,
    Unindent(`
      file { '/tmp/foo':
        'mode' => '0640',
        'ensure' => present
      }`),
    `expected attribute name at line 2:3`)
}

func TestMultipleBodies(t *testing.T) {
  expectDump(t,
    Unindent(`
      file { '/tmp/foo':
        mode => '0640',
        ensure => present;
      '/tmp/bar':
        mode => '0640',
        ensure => present;
      }`),
    `(resource {:type (qn file) :bodies [`+
        `{:title "/tmp/foo" :ops [(=> (mode) "0640") (=> (ensure) (qn present))]} `+
        `{:title "/tmp/bar" :ops [(=> (mode) "0640") (=> (ensure) (qn present))]}]})`)

  expectError(t,
    Unindent(`
      file { '/tmp/foo':
        mode => '0640',
        ensure => present;
      '/tmp/bar'
        mode => '0640',
        ensure => present;
      }`),
    `resource title expected at line 4:1`)
}


func TestStatmentCallWithUnparameterizedHash(t *testing.T) {
  expectDump(t,
    `warning { message => 'syntax ok' }`,
    `(invoke {:functor (qn warning) :args [(hash (=> (qn message) "syntax ok"))]})`)
}

func TestNonStatmentCallWithUnparameterizedHash(t *testing.T) {
  expectError(t,
    `something { message => 'syntax ok' }`,
    `This expression is invalid. Did you try declaring a 'something' resource without a title? at line 1:1`)
}

func TestResourceDefaults(t *testing.T) {
  expectDump(t,
    `Something { message => 'syntax ok' }`,
    `(resource-defaults {:type (qr Something) :ops [(=> (message) "syntax ok")]})`)
}

func TestResourceDefaultsFromAccess(t *testing.T) {
  expectDump(t,
    `Resource[Something] { message => 'syntax ok' }`,
    `(resource-defaults {:type ([] (qr Resource) (qr Something)) :ops [(=> (message) "syntax ok")]})`)

  expectDump(t,
    `@Resource[Something] { message => 'syntax ok' }`,
    `(resource-defaults {:form virtual :type ([] (qr Resource) (qr Something)) :ops [(=> (message) "syntax ok")]})`)
}

func TestResourceOverride(t *testing.T) {
  expectDump(t,
    `File['/tmp/foo.txt'] { mode => '0644' }`,
    `(resource-override {:resources ([] (qr File) "/tmp/foo.txt") :ops [(=> (mode) "0644")]})`)

  expectDump(t,
    Unindent(`
      Service['apache'] {
        require +> [File['apache.pem'], File['httpd.conf']]
      }`),
    `(resource-override {:resources ([] (qr Service) "apache") :ops [(+> (require) (array ([] (qr File) "apache.pem") ([] (qr File) "httpd.conf")))]})`)

  expectDump(t,
    `@File['/tmp/foo.txt'] { mode => '0644' }`,
    `(resource-override {:form virtual :resources ([] (qr File) "/tmp/foo.txt") :ops [(=> (mode) "0644")]})`)

}

func TestInvalidResource(t *testing.T) {
  expectError(t,
    `'File' { mode => '0644' }`,
    `invalid resource expression at line 1:1`)
}

func TestVirtualResourceCollector(t *testing.T) {
  expectDump(t,
    `File <| |>`,
    `(collect {:type (qr File) :query (<| |>)})`)

  expectDump(t,
    Unindent(`
      File <| mode == '0644' |>`),
    `(collect {:type (qr File) :query (<| |> (== (qn mode) "0644"))})`)

  expectDump(t,
    Unindent(`
      File <| mode == '0644' |> {
        owner => 'root',
        mode => 640
      }`),
    `(collect {:type (qr File) :query (<| |> (== (qn mode) "0644")) :ops [(=> (owner) "root") (=> (mode) 640)]})`)
}

func TestExportedResourceCollector(t *testing.T) {
  expectDump(t,
    `File <<| |>>`,
    `(collect {:type (qr File) :query (<<| |>>)})`)

  expectDump(t,
    Unindent(`
      File <<| mode == '0644' |>>`),
    `(collect {:type (qr File) :query (<<| |>> (== (qn mode) "0644"))})`)

  expectDump(t,
    Unindent(`
      File <<| mode == '0644' |>> {
        owner => 'root',
        mode => 640
      }`),
    `(collect {:type (qr File) :query (<<| |>> (== (qn mode) "0644")) :ops [(=> (owner) "root") (=> (mode) 640)]})`)
}

func TestOperators(t *testing.T) {
  expectDump(t,
    `$x = a or b and c < d == e << f + g * -h`,
    `(= ($ x) (or (qn a) (and (qn b) (< (qn c) (== (qn d) (<< (qn e) (+ (qn f) (* (qn g) (- (qn h))))))))))`)

  expectDump(t,
    `$x = -h / g + f << e == d <= c and b or a`,
    `(= ($ x) (or (and (<= (== (<< (+ (/ (- (qn h)) (qn g)) (qn f)) (qn e)) (qn d)) (qn c)) (qn b)) (qn a)))`)

  expectDump(t,
    `$x = !a == b`,
    `(= ($ x) (== (! (qn a)) (qn b)))`)

  expectDump(t,
    `$x = a > b`,
    `(= ($ x) (> (qn a) (qn b)))`)

  expectDump(t,
    `$x = a >= b`,
    `(= ($ x) (>= (qn a) (qn b)))`)

  expectDump(t,
    `$x = a +b`,
    `(= ($ x) (+ (qn a) (qn b)))`)

  expectDump(t,
    `$x = +4`,
    `(= ($ x) 4)`)

  expectDump(t,
    `$x = a * (b + c)`,
    `(= ($ x) (* (qn a) (() (+ (qn b) (qn c)))))`)

  expectDump(t,
    `$x = $y -= $z`,
    `(= ($ x) (-= ($ y) ($ z)))`)

  expectDump(t,
    `$x = $y + $z % 5`,
    `(= ($ x) (+ ($ y) (% ($ z) 5)))`)

  expectDump(t,
    `$x = $y += $z`,
    `(= ($ x) (+= ($ y) ($ z)))`)

  expectError(t,
    `$x = +b`,
    `unexpected token '+' at line 1:7`)
}

func TestMatch(t *testing.T) {
  expectDump(t,
    `a =~ /^[a-z]+$/`,
    `(=~ (qn a) (regexp "^[a-z]+$"))`)

  expectDump(t,
    `a !~ /^[a-z]+$/`,
    `(!~ (qn a) (regexp "^[a-z]+$"))`)
}

func TestIn(t *testing.T) {
  expectDump(t,
    `'eat' in 'eaten'`,
    `(in "eat" "eaten")`)

  expectDump(t,
    `'eat' in ['eat', 'ate', 'eating']`,
    `(in "eat" (array "eat" "ate" "eating"))`)
}


func dump(e Expression) string {
  result := bytes.NewBufferString(``)
  e.ToPN().Format(result)
  return result.String()
}

func TestEPP(t *testing.T) {
  expectDumpEPP(t,
    ``,
    `(render-s "")`)

  expectDumpEPP(t,
    Unindent(`
      some arbitrary text
      spanning multiple lines`),
    `(render-s "some arbitrary text\nspanning multiple lines")`)

  expectDumpEPP(t,
    Unindent(`
      <%||%> some arbitrary text
      spanning multiple lines`),
    `(lambda {:body [(render-s " some arbitrary text\nspanning multiple lines")]})`)

  expectDumpEPP(t,
    Unindent(`
      <%||%> some <%#-%>text`),
    `(lambda {:body [(render-s " some text")]})`)

  expectErrorEPP(t,
    Unindent(`
      <%||%> some <%#-text`),
    `unbalanced epp comment at line 1:13`)

  expectDumpEPP(t,
    Unindent(`
      <%||%> some <%%-%%-%%> text`),
    `(lambda {:body [(render-s " some <%-%%-%> text")]})`)

  expectDumpEPP(t,
    Unindent(`
      <%||-%> some <-% %-> text`),
    `(lambda {:body [(render-s "some <-% %-> text")]})`)

  expectDumpEPP(t,
    Unindent(`
      <%-||-%> some <%- $x = 3 %> text`),
    `(lambda {:body [(render-s "some") (= ($ x) 3) (render-s " text")]})`)

  expectErrorEPP(t,
    Unindent(`
      <%-||-%> some <%- $x = 3 -% $y %> text`),
    `invalid operator '-%' at line 1:28`)

  expectBlockEPP(t,
    Unindent(`
      vcenter: {
        host: "<%= $host %>"
        user: "<%= $username %>"
        password: "<%= $password %>"
      }`),
    `(block `+
        `(render-s "vcenter: {\n  host: \"") `+
        `(render ($ host)) `+
        `(render-s "\"\n  user: \"") `+
        `(render ($ username)) `+
        `(render-s "\"\n  password: \"") `+
        `(render ($ password)) `+
        `(render-s "\"\n}"))`)

  expectDumpEPP(t,
    Unindent(`
      <%- | Boolean $keys_enable,
        String  $keys_file,
        Array   $keys_trusted,
        String  $keys_requestkey,
        String  $keys_controlkey
      | -%>
      <%# Parameter tag ↑ -%>

      <%# Non-printing tag ↓ -%>
      <% if $keys_enable { -%>

      <%# Expression-printing tag ↓ -%>
      keys <%= $keys_file %>
      <% unless $keys_trusted =~ Array[Data,0,0] { -%>
      trustedkey <%= $keys_trusted.join(' ') %>
      <% } -%>
      <% if $keys_requestkey =~ String[1] { -%>
      requestkey <%= $keys_requestkey %>
      <% } -%>
      <% if $keys_controlkey =~ String[1] { -%>
      controlkey <%= $keys_controlkey %>
      <% } -%>

      <% } -%>`),
    `(lambda {`+
      `:params [`+
        `{:name keys_enable :type (qr Boolean)} `+
        `{:name keys_file :type (qr String)} `+
        `{:name keys_trusted :type (qr Array)} `+
        `{:name keys_requestkey :type (qr String)} `+
        `{:name keys_controlkey :type (qr String)}] `+
      `:body [`+
        `(render-s "\n\n\n") `+
        `(if {`+
          `:test ($ keys_enable) `+
          `:then [(render-s "\n\nkeys ") `+
            `(render ($ keys_file)) `+
            `(render-s "\n") `+
            `(unless {`+
              `:test (=~ ($ keys_trusted) ([] (qr Array) (qr Data) 0 0)) `+
              `:then [`+
                `(render-s "trustedkey ") `+
                `(render (call_method {:functor (. ($ keys_trusted) (qn join)) :args [" "]})) `+
                `(render-s "\n")]}) `+
            `(if {`+
              `:test (=~ ($ keys_requestkey) ([] (qr String) 1)) `+
              `:then [`+
                `(render-s "requestkey ") `+
                `(render ($ keys_requestkey)) `+
                `(render-s "\n")]}) `+
            `(if {`+
              `:test (=~ ($ keys_controlkey) ([] (qr String) 1)) `+
              `:then [`+
                `(render-s "controlkey ") `+
                `(render ($ keys_controlkey)) `+
                `(render-s "\n")]}) `+
                `(render-s "\n")]})]})`)

  // Fail on EPP constructs unless EPP is enabled
  expectError(t,
    Unindent(`
      <% $x = 3 %> text`),
    `unexpected token '<' at line 1:1`)

  expectError(t,
    Unindent(`
      $x = 3 %> 4`),
    `unexpected token '>' at line 1:9`)

  expectError(t,
    Unindent(`
      $x = 3 -%> 4`),
    `unexpected token '%' at line 1:9`)

  expectErrorEPP(t,
    "\n<% |String $x| %>\n",
    `Ambiguous EPP parameter expression. Probably missing '<%-' before parameters to remove leading whitespace at line 2:5`)
}

func expectDump(t *testing.T, source string, expected string) {
  expectDumpX(t, source, expected, false)
}

func expectDumpEPP(t *testing.T, source string, expected string) {
  expectDumpX(t, source, expected, true)
}

func expectBlockEPP(t *testing.T, source string, expected string) {
  expectBlockX(t, source, expected, true)
}

func expectDumpX(t *testing.T, source string, expected string, eppMode bool) {
  if expr := parseExpression(t, source, eppMode); expr != nil {
    actual := dump(expr)
    if expected != actual {
      t.Errorf("expected '%s', got '%s'", expected, actual)
    }
  }
}

func expectBlock(t *testing.T, source string, expected string) {
  expectBlockX(t, source, expected, false)
}

func expectBlockX(t *testing.T, source string, expected string, eppMode bool) {
  expr, err := Parse(``, source, eppMode)
  if err != nil {
    t.Errorf(err.Error())
  } else {
    actual := dump(expr)
    if expected != actual {
      t.Errorf("expected '%s', got '%s'", expected, actual)
    }
  }
}

func expectError(t *testing.T, source string, expected string) {
  expectErrorX(t, source, expected, false)
}

func expectErrorEPP(t *testing.T, source string, expected string) {
  expectErrorX(t, source, expected, true)
}

func expectErrorX(t *testing.T, source string, expected string, eppMode bool) {
  _, err := Parse(``, source, eppMode)
  if err == nil {
    t.Errorf("Expected error '%s' but nothing was raised", expected)
  } else {
    actual := err.Error()
    if expected != actual {
      t.Errorf("expected error '%s', got '%s'", expected, actual)
    }
  }
}

func expectHeredoc(t *testing.T, str string, args ...interface{}) {
  expected := args[0].(string)
  expr := parseExpression(t, str, false)
  if expr == nil {
    return
  }
  if heredoc, ok := expr.(*HeredocExpression); ok {
    if len(args) > 1 && heredoc.syntax != args[1] {
      t.Errorf("Expected syntax '%s', got '%s'", args[1], heredoc.syntax)
    }
    if textExpr, ok := heredoc.text.(*LiteralString); ok {
      if textExpr.value != expected {
        t.Errorf("Expected heredoc '%s', got '%s'", expected, textExpr.value)
      }
      return
    }
    actual := dump(expr)
    if actual != expected {
      t.Errorf("Expected heredoc '%s', got '%s'", expected, actual)
    }
    return
  }
  t.Errorf("'%s' did not result in a heredoc expression", str)
}

func parse(t *testing.T, str string, eppMode bool) Expression {
  expr, err := Parse(``, str, eppMode)
  if err != nil {
    t.Errorf(err.Error())
    return nil
  }
  program, ok := expr.(*Program)
  if !ok {
    t.Errorf("'%s' did not parse to a program", str)
    return nil
  }
  return program.body
}

func parseExpression(t *testing.T, str string, eppMode bool) Expression {
  expr := parse(t, str, eppMode)
  if block, ok := expr.(*BlockExpression); ok {
    if len(block.statements) == 1 {
      return block.statements[0]
    }
    t.Errorf("'%s' did not parse to a block with exactly one expression", str)
    return nil
  }
  return expr
}