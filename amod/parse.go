package amod

import (
	"io"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Use participle to parse the lexemes.
// https://github.com/alecthomas/participle

type amodFile struct {
	Model       *modelSection      `parser:"'==model==' @@"`
	Config      *configSection     `parser:"'==config==' (@@)?"`
	Init        *initSection       `parser:"('==init==' (@@)?)?"`
	Productions *productionSection `parser:"'==productions==' (@@)?"`

	Pos lexer.Position
}

type modelSection struct {
	Name        string   `parser:"'name' ':' (@String|@Ident)"`
	Description string   `parser:"('description' ':' (@String|@Ident))?"`
	Examples    []string `parser:"('examples' '{' (@String)+ '}')?"`

	Pos lexer.Position
}

type identList struct {
	Identifiers []string `parser:"( @Ident ','? )+"`

	Pos lexer.Position
}

type stringList struct {
	Strings []string `parser:"( @String ','? )+"`

	Pos lexer.Position
}

type arg struct {
	Arg string `parser:"@String|@Ident|@Number"`

	Pos lexer.Position
}

// String returns just the string portion of an arg struct
func (a *arg) String() string {
	return a.Arg
}

type argList struct {
	Args []*arg `parser:"( @@ ','? )+"`

	Pos lexer.Position
}

// Strings converts an arg list into a string slice
func (a *argList) Strings() []string {
	strs := make([]string, len(a.Args))
	for i, arg := range a.Args {
		strs[i] = arg.Arg
	}

	return strs
}

type value struct {
	String *string  `parser:"  (@String|@Ident)"`
	Number *float64 `parser:"| @Number"`

	Pos lexer.Position
}

type field struct {
	Key   string `parser:"@Ident ':'"`
	Value value  `parser:"@@ (',')?"`

	Pos lexer.Position
}

type fieldList struct {
	Fields []*field `parser:"'{' @@+ '}'"`

	Pos lexer.Position
}

type memory struct {
	Name   string    `parser:"@Ident"`
	Fields fieldList `parser:"@@+"`

	Pos lexer.Position
}

type memoryList struct {
	Memories []*memory `parser:"'{' @@+ '}'"`

	Pos lexer.Position
}

type configSection struct {
	ACTR        *fieldList  `parser:"('actr' @@)?"`
	Buffers     *identList  `parser:"('buffers' '{' @@ '}')?"`
	Memories    *memoryList `parser:"('memories' @@)?"`
	TextOutputs *identList  `parser:"('text_outputs' '{' @@ '}')?"`

	Pos lexer.Position
}

type initializer struct {
	Name  string      `parser:"@Ident"`
	Items *stringList `parser:"'{' @@+ '}'"`

	Pos lexer.Position
}

type initSection struct {
	Initializers []*initializer `parser:"@@+"`

	Pos lexer.Position
}

type patternFieldItem struct {
	ID             *string `parser:"( @Ident"`
	Num            *string `parser:"| @Number"` // we don't need to treat this as a number anywhere, so keep as a string
	Var            *string `parser:"| @PatternVar"`
	NotVar         *string `parser:"| '!' @PatternVar"`
	OptionalVar    *string `parser:"| '?' @PatternVar"`
	NotOptionalVar *string `parser:"| '!' '?' @PatternVar)"`
}

type patternField struct {
	Name  *string             `parser:"(@Ident ':')?"`
	Items []*patternFieldItem `parser:"@@+"`
	Space string              `parser:" @PatternSpace? "`

	Pos lexer.Position
}

type pattern struct {
	Fields []*patternField "parser:\"'`' @@+ '`'\""

	Pos lexer.Position
}

type matchItem struct {
	Name    string   `parser:"@Ident ':'"`
	Text    *string  `parser:"(@String"`
	Pattern *pattern `parser:"| @@)"`

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
	Args *argList `parser:"'print' @@?"`

	Pos lexer.Position
}

type recallStatement struct {
	Pattern    *pattern `parser:"'recall' @@"`
	MemoryName string   `parser:"'from' @Ident"`

	Pos lexer.Position
}

type writeStatement struct {
	Args           *argList `parser:"'write' @@"`
	TextOutputName string   `parser:"'to' @Ident"`

	Pos lexer.Position
}

type setField struct {
	ArgNum *float64 `parser:"'field' (@Number"`
	Name   *string  `parser:"| @Ident)"`

	Pos lexer.Position
}

type setStatement struct {
	Set        string    `parser:"'set'"` // not used, but must be visible for parse to work
	Field      *setField `parser:"(@@ 'of')?"`
	BufferName string    `parser:"@Ident"`
	Arg        *arg      `parser:"'to' (@@"`
	Pattern    *pattern  `parser:"| @@)"`

	Pos lexer.Position
}

type statement struct {
	Clear  *clearStatement  `parser:"  @@"`
	Print  *printStatement  `parser:"| @@"`
	Recall *recallStatement `parser:"| @@"`
	Set    *setStatement    `parser:"| @@"`
	Write  *writeStatement  `parser:"| @@"`

	Pos lexer.Position
}

type do struct {
	Do         string        `parser:"'do'"` // not used, but must be visible for parse to work
	PyCode     *[]string     `parser:"('#<' (@DoCode)+ '>#'"`
	Statements *[]*statement `parser:"| '{' (@@)+ '}')"`

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

var parser = participle.MustBuild(&amodFile{},
	participle.Lexer(LexerDefinition),
	participle.Elide("Comment", "Whitespace"),
)

func parse(r io.Reader) (*amodFile, error) {
	var amod amodFile

	err := parser.Parse("", r, &amod)

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

func (p patternFieldItem) getVar() *string {
	if p.Var != nil {
		return p.Var
	} else if p.NotVar != nil {
		return p.NotVar
	} else if p.OptionalVar != nil {
		return p.OptionalVar
	} else if p.NotOptionalVar != nil {
		return p.NotOptionalVar
	}

	return nil
}
