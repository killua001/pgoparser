package tokenizer

import (
	"github.com/elliotcourant/pgoparser/tokens"
)

var (
	eof = byte(0)
)

type (
	Tokenizer struct {
		input  string
		offset int
	}
)

func (t *Tokenizer) peak() byte {
	if len(t.input) < t.offset+1 {
		return eof
	}

	return t.input[t.offset]
}

func (t *Tokenizer) scan() byte {
	if len(t.input) < t.offset+1 {
		return eof
	}

	t.offset++

	return t.input[t.offset-1]
}

func (t *Tokenizer) nextToken() tokens.Token {
	character := t.peak()
	switch character {
	case ' ':
		panic("whitespace not implemented")
	case '\t':
		panic("whitespace not implemented")
	case '\n', 'r':
		panic("whitespace not implemented")
	case '\'':
		panic("single quoted string not implemented")
	case '"':
		panic("double quoted string not implemented")
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		panic("numbers not implemented")
	case '(':

	}

	return nil
}
