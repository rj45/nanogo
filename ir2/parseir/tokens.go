package parseir

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF           // end of file
	WS            // whitespace
	NL            // newline

	// Literals
	IDENT
	NUM
	STR

	// Misc characters
	ASTERISK // *
	COMMA    // ,
	COLON    // :
	EQUALS   // =
	DOT      // .

	// Keywords
	FUNC
	PACKAGE
)
