package mayo

import (
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenIdent tokenType = iota
	tokenHash
	tokenString
	tokenNumber
	tokenDimension
	tokenPercentage
	tokenWhitespace
	tokenColon
	tokenSemicolon
	tokenComma
	tokenOpenBrace
	tokenCloseBrace
	tokenOpenParen
	tokenCloseParen
	tokenOpenBracket
	tokenCloseBracket
	tokenDelim
	tokenAtKeyword
	tokenEOF
	tokenGT
	tokenPlus
	tokenTilde
	tokenDot
	tokenStar
)

type cssToken struct {
	Type  tokenType
	Value string
}

type cssTokenizer struct {
	input []rune
	pos   int
}

func newTokenizer(input string) *cssTokenizer {
	return &cssTokenizer{
		input: []rune(input),
		pos:   0,
	}
}

func (t *cssTokenizer) peek() rune {
	if t.pos >= len(t.input) {
		return 0
	}
	return t.input[t.pos]
}

func (t *cssTokenizer) advance() rune {
	if t.pos >= len(t.input) {
		return 0
	}
	ch := t.input[t.pos]
	t.pos++
	return ch
}

func (t *cssTokenizer) skipWhitespace() string {
	start := t.pos
	for t.pos < len(t.input) && unicode.IsSpace(t.input[t.pos]) {
		t.pos++
	}
	return string(t.input[start:t.pos])
}

func (t *cssTokenizer) skipComment() bool {
	if t.pos+1 < len(t.input) && t.input[t.pos] == '/' && t.input[t.pos+1] == '*' {
		t.pos += 2
		for t.pos+1 < len(t.input) {
			if t.input[t.pos] == '*' && t.input[t.pos+1] == '/' {
				t.pos += 2
				return true
			}
			t.pos++
		}
		t.pos = len(t.input)
		return true
	}
	return false
}

func (t *cssTokenizer) readIdent() string {
	var sb strings.Builder
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '_' {
			sb.WriteRune(ch)
			t.pos++
		} else if ch == '\\' && t.pos+1 < len(t.input) {
			t.pos++
			sb.WriteRune(t.input[t.pos])
			t.pos++
		} else {
			break
		}
	}
	return sb.String()
}

func (t *cssTokenizer) readString(quote rune) string {
	var sb strings.Builder
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == quote {
			t.pos++
			break
		}
		if ch == '\\' && t.pos+1 < len(t.input) {
			t.pos++
			sb.WriteRune(t.input[t.pos])
		} else {
			sb.WriteRune(ch)
		}
		t.pos++
	}
	return sb.String()
}

func (t *cssTokenizer) readNumber() string {
	var sb strings.Builder
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if unicode.IsDigit(ch) || ch == '.' || ch == '-' || ch == '+' {
			sb.WriteRune(ch)
			t.pos++
		} else {
			break
		}
	}
	return sb.String()
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '-'
}

func (t *cssTokenizer) tokenize() []cssToken {
	var tokens []cssToken

	for t.pos < len(t.input) {
		// Skip comments
		if t.skipComment() {
			continue
		}

		ch := t.peek()

		// Whitespace
		if unicode.IsSpace(ch) {
			ws := t.skipWhitespace()
			tokens = append(tokens, cssToken{tokenWhitespace, ws})
			continue
		}

		// String
		if ch == '"' || ch == '\'' {
			t.advance()
			str := t.readString(ch)
			tokens = append(tokens, cssToken{tokenString, str})
			continue
		}

		// Hash
		if ch == '#' {
			t.advance()
			ident := t.readIdent()
			tokens = append(tokens, cssToken{tokenHash, "#" + ident})
			continue
		}

		// At-keyword
		if ch == '@' {
			t.advance()
			ident := t.readIdent()
			tokens = append(tokens, cssToken{tokenAtKeyword, "@" + ident})
			continue
		}

		// Number
		if unicode.IsDigit(ch) || (ch == '.' && t.pos+1 < len(t.input) && unicode.IsDigit(t.input[t.pos+1])) {
			num := t.readNumber()
			// Check for dimension or percentage
			if t.pos < len(t.input) && t.input[t.pos] == '%' {
				t.advance()
				tokens = append(tokens, cssToken{tokenPercentage, num + "%"})
			} else if t.pos < len(t.input) && isIdentStart(t.input[t.pos]) {
				unit := t.readIdent()
				tokens = append(tokens, cssToken{tokenDimension, num + unit})
			} else {
				tokens = append(tokens, cssToken{tokenNumber, num})
			}
			continue
		}

		// Ident
		if isIdentStart(ch) {
			ident := t.readIdent()
			tokens = append(tokens, cssToken{tokenIdent, ident})
			continue
		}

		// Single-character tokens
		t.advance()
		switch ch {
		case '{':
			tokens = append(tokens, cssToken{tokenOpenBrace, "{"})
		case '}':
			tokens = append(tokens, cssToken{tokenCloseBrace, "}"})
		case '(':
			tokens = append(tokens, cssToken{tokenOpenParen, "("})
		case ')':
			tokens = append(tokens, cssToken{tokenCloseParen, ")"})
		case '[':
			tokens = append(tokens, cssToken{tokenOpenBracket, "["})
		case ']':
			tokens = append(tokens, cssToken{tokenCloseBracket, "]"})
		case ':':
			tokens = append(tokens, cssToken{tokenColon, ":"})
		case ';':
			tokens = append(tokens, cssToken{tokenSemicolon, ";"})
		case ',':
			tokens = append(tokens, cssToken{tokenComma, ","})
		case '>':
			tokens = append(tokens, cssToken{tokenGT, ">"})
		case '+':
			tokens = append(tokens, cssToken{tokenPlus, "+"})
		case '~':
			tokens = append(tokens, cssToken{tokenTilde, "~"})
		case '.':
			tokens = append(tokens, cssToken{tokenDot, "."})
		case '*':
			tokens = append(tokens, cssToken{tokenStar, "*"})
		default:
			tokens = append(tokens, cssToken{tokenDelim, string(ch)})
		}
	}

	tokens = append(tokens, cssToken{tokenEOF, ""})
	return tokens
}
