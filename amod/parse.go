package amod

import (
	"io"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Uses participle to parse the lexemes.
// 	https://github.com/alecthomas/participle

// Railroad Diagrams
// ------
// First output the EBNF grammar to stdout with the command "gactar -ebnf".
//
// There are two ways to generate railroad diagrams:
// 	1. Use the "railroad" tool from participle like this:
//		./railroad -o amod-grammar.html -w
//		paste in the generated EBNF above & hit control-D
//	2. Use this page to convert the ebnf and generate a diagram:
//		https://bottlecaps.de/convert/
//		paste in the generated EBNF above, click "Convert" and then click "View Diagram"

type amodFile struct {
	Model       *modelSection      `parser:"'==model==' @@"`
	Config      *configSection     `parser:"'==config==' (@@)?"`
	Init        *initSection       `parser:"('==init==' (@@)?)?"`
	Productions *productionSection `parser:"'==productions==' (@@)?"`

	Pos lexer.Position
}

type modelSection struct {
	Name        string     `parser:"'name' ':' (@String|@Ident)"`
	Description string     `parser:"('description' ':' (@String|@Ident))?"`
	Examples    []*pattern `parser:"('examples' '{' @@+ '}')?"`

	Pos lexer.Position
}

type arg struct {
	Var    *string `parser:"( @PatternVar"`
	ID     *string `parser:"| @Ident"`
	Str    *string `parser:"| @String"`
	Number *string `parser:"| @Number)"`

	Pos lexer.Position
}

type fieldValue struct {
	ID     *string  `parser:"( @Ident"`
	Str    *string  `parser:"| @String"`
	Number *float64 `parser:"| @Number)"`

	Pos lexer.Position
}

type field struct {
	Key   string     `parser:"@Ident ':'"`
	Value fieldValue `parser:"@@ (',')?"`

	Pos lexer.Position
}

type chunkDecl struct {
	Name      string   `parser:"@Ident"`
	SlotNames []string `parser:"'(' ( @Ident ','? )+ ')'"`

	Pos lexer.Position
}

type configSection struct {
	ACTR       []*field     `parser:"('actr' '{' @@+ '}')?"`
	ChunkDecls []*chunkDecl `parser:"('chunks' '{' @@+ '}')?"`
	MemoryDecl []*field     `parser:"('memory' '{' @@+ '}')?"`

	Pos lexer.Position
}

type initSection struct {
	Patterns []*pattern `parser:"'memory' '{' @@+ '}'"`

	Pos lexer.Position
}

type patternSlotItem struct {
	Nil    *bool   `parser:"( @('nil':Keyword)"`
	ID     *string `parser:"| @Ident"`
	Num    *string `parser:"| @Number"` // we don't need to treat this as a number anywhere, so keep as a string
	Var    *string `parser:"| @PatternVar"`
	NotVar *string `parser:"| '!' @PatternVar)"`

	Pos lexer.Position
}

type patternSlot struct {
	Space1 string             `parser:" @PatternSpace? "`
	Items  []*patternSlotItem `parser:"@@+"`
	Space2 string             `parser:" @PatternSpace? "`

	Pos lexer.Position
}

type pattern struct {
	StartTick string         "parser:\"'`'\"" // not used - must be set for parse
	ChunkName string         `parser:" @Ident '('"`
	Slots     []*patternSlot `parser:" @@+ ')'"`
	EndTick   string         "parser:\"'`'\"" // not used - must be set for parse

	Pos lexer.Position
}

type matchItem struct {
	Name    string   `parser:"(@Ident|@('memory':Keyword))"`
	Pattern *pattern `parser:" @@ "`

	Pos lexer.Position
}

type match struct {
	Items []*matchItem `parser:"'match' '{' @@+ '}'"`

	Pos lexer.Position
}

type clearStatement struct {
	BufferNames []string `parser:"'clear' ( @Ident ','? )+"`

	Pos lexer.Position
}

type printStatement struct {
	Args []*arg `parser:"'print' ( @@ ','? )*"`

	Pos lexer.Position
}

type recallStatement struct {
	Pattern *pattern `parser:"'recall' @@"`

	Pos lexer.Position
}

type setValue struct {
	Nil    *bool   `parser:"( @('nil':Keyword)"`
	Var    *string `parser:"| @PatternVar"`
	Str    *string `parser:"| @String"`
	Number *string `parser:"| @Number)"`

	Pos lexer.Position
}

type setStatement struct {
	Set        string  `parser:"'set'"` // not used, but must be visible for parse to work
	BufferName string  `parser:"@Ident"`
	Slot       *string `parser:"('.' @Ident)?"`

	Value   *setValue `parser:"'to' (@@"`
	Pattern *pattern  `parser:"| @@)"`

	Pos lexer.Position
}

type statement struct {
	Clear  *clearStatement  `parser:"  @@"`
	Print  *printStatement  `parser:"| @@"`
	Recall *recallStatement `parser:"| @@"`
	Set    *setStatement    `parser:"| @@"`

	Pos lexer.Position
}

type do struct {
	Do         string        `parser:"'do'"` // not used, but must be visible for parse to work
	Statements *[]*statement `parser:"'{' @@+ '}'"`

	Pos lexer.Position
}

type production struct {
	Name  string `parser:"@Ident '{'"`
	Match *match `parser:"@@"`
	Do    *do    `parser:"@@"`
	End   string `parser:"'}'"` // not used, but must be visible for parse to work

	Pos lexer.Position
}

type productionSection struct {
	Productions []*production `parser:"( @@ )+"`

	Pos lexer.Position
}

var amodParser = participle.MustBuild(&amodFile{},
	participle.Lexer(LexerDefinition),
	participle.Elide("Comment", "Whitespace"),
	participle.Unquote(),
)

var patternParser = participle.MustBuild(&pattern{},
	participle.Lexer(LexerDefinition),
	participle.Elide("Comment", "Whitespace"),
	participle.Unquote(),
)

func parse(r io.Reader) (*amodFile, error) {
	var amod amodFile

	err := amodParser.Parse("", r, &amod)

	if err != nil {
		return nil, err
	}

	return &amod, nil
}

func parseFile(filename string) (*amodFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parse(file)
}

func (p patternSlotItem) getVar() *string {
	if p.Var != nil {
		return p.Var
	} else if p.NotVar != nil {
		return p.NotVar
	}

	return nil
}
