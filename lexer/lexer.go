package lexer

import (
	"fmt"
	"solparsor/token"
	"strings"
	"unicode/utf8"
)

const (
	eof        = 0
	contract   = "Contract"
	leftBrace  = "{"
	rightBrace = "}"
)

// The state represents where we are in the input and what we expect to see next.
// An action defines what we are going to do in that state given the input.
// After you execute the action, you will be in a new state.
// Combining the state and the action together results in a state function.
// The stateFn represents the state of the lexer as a function that returns the next state.
// It is a recursive definition.
type stateFn func(*lexer) stateFn

// The `run` function lexes the input by executing state functions
// until the state is nil.
func (l *lexer) run() {
	// The initial state is lexSourceUnit. SourceUnit is basically a Solidity file.
	for state := lexSourceUnit; state != nil; {
		state = state(l)
	}
	// The lexer is done, so we close the channel.
	// It tells the caller (probably the parser),
	// that no more tokens will be delivered.
	close(l.tokens)
}

// The lexer holds the state of the scanner.
type lexer struct {
	input  string           // The string being scanned.
	start  int              // Start position of this token.Token; in a big string, this is the start of the current token.
	pos    int              // Current position in the input.
	width  int              // Width of last rune read from input.
	tokens chan token.Token // Channel of scanned token.
}

func Lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token.Token, 2), // Buffer 2 tokens. We don't need more.
	}
	println("Lexing input: ", input)
	fmt.Printf("Input length: %d\n\n", len(input))
	// This starts the state machine.
	go l.run()

	return l
}

func (l *lexer) NextToken() token.Token {
	for {
		select {
		case tkn := <-l.tokens:
			return tkn
		}
	}
}

// The `emit` function passes an token.Token back to the client.
func (l *lexer) emit(typ token.TokenType) {
	println("Emitting: ", l.input[l.start:l.pos])
	// The value is a slice of the input.
	l.tokens <- token.Token{
		Type:    typ,
		Literal: l.input[l.start:l.pos],
		Pos:     token.Position(l.start),
	}
	// Move ahead in the input after sending it to the caller.
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{
		Type:    token.ILLEGAL,
		Literal: fmt.Sprintf(format, args...),
		Pos:     token.Position(l.start),
	}
	return nil
}

func lexSourceUnit(l *lexer) stateFn {
	for {
		switch char := l.readChar(); {
		case char == eof:
			l.emit(token.EOF)
			return nil
		case isWhitespace(char):
			l.ignore()
		case isLetter(char):
			l.backup()
			return lexIdentifier
		case isDigit(char):
			l.backup()
			return lexNumber
		case char == ';':
			l.emit(token.SEMICOLON)
		case char == '{':
			l.emit(token.LBRACE)
		case char == '}':
			l.emit(token.RBRACE)
		case char == '(':
			l.emit(token.LPAREN)
		case char == ')':
			l.emit(token.RPAREN)
		case char == '[':
			l.emit(token.LBRACKET)
		case char == ']':
			l.emit(token.RBRACKET)
		case char == '.':
			l.emit(token.PERIOD)
		case char == '!':
			l.emit(l.switch2(token.NOT, token.NOT_EQUAL))
		case char == '=':
			l.emit(l.switch3(token.ASSIGN, token.EQUAL, ">", token.DOUBLE_ARROW))
		case char == '+':
			l.emit(l.switch3(token.ADD, token.ASSIGN_ADD, "+", token.INC))
		case char == '-':
			if l.accept(">") {
				l.emit(token.RIGHT_ARROW)
				continue
			}
			l.emit(l.switch3(token.SUB, token.ASSIGN_SUB, "-", token.DEC))
		case char == '<':
			l.emit(l.switch4(
				token.LESS_THAN, token.LESS_THAN_OR_EQUAL, "<",
				token.SHL, token.ASSIGN_SHL))
		case char == '>':
			// There are 6 cases for the '>' character. We handle the '>=' and '>'
			// separately. The remaining 4 cases are handled by the switch4 helper.
			tkn := token.GREATER_THAN
			if l.accept("=") {
				tkn = token.GREATER_THAN_OR_EQUAL
			} else if l.accept(">") {
				tkn = l.switch4(token.SAR, token.ASSIGN_SAR, ">", token.SHR, token.ASSIGN_SHR)
			}
			l.emit(tkn)
		default:
			return l.errorf("Unrecognised character in source unit: '%c'", char)
		}
	}
}

func lexIdentifier(l *lexer) stateFn {
	for {
		switch char := l.readChar(); {
		case isLetter(char):
			// Do nothing, just keep reading.
		case isDigit(char):
			// Do nothing, just keep reading.
			// We entered here so we know that the first char is a letter.
			// We can have digits after letters in the identifiers.
		default:
			// We are sitting on something different than alphanumeric so just go back.
			l.backup()
			l.emit(token.LookupIdent(l.input[l.start:l.pos]))
			return lexSourceUnit
		}
	}
}

func lexNumber(l *lexer) stateFn {
	hex := false
	l.accept("+-") // The sign is optional.
	digits := "0123456789"

	// Is the number hexadecimal? Starts with 0x?
	if l.accept("0") && l.accept("x") {
		// If so, we need to extend the valid set of digits.
		digits = "0123456789abcdefABCDEF"
		hex = true
	}

	l.acceptRun(digits)

	// @TODO: Fixed point numbers could probably go here. Solidity have them,
	// but you can't use them yet, soooo...

	// Does it have an exponent at the end? For example: 100e10 or 1000000e-3.
	// Solidity allows both `e` and `E` as the exponent.
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789") // Hex is not allowed in the exponent.
	}

	if hex {
		l.emit(token.HEX_NUMBER)
	} else {
		l.emit(token.DECIMAL_NUMBER)
	}
	return lexSourceUnit
}

// readChar reads the next rune from the input, advances the position
// and returns the rune.
func (l *lexer) readChar() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width

	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.readChar()
	l.backup()
	return r
}

// accept consumes the next rune if it's from the valid set. If not, it backs up.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.readChar()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes runes as long as they are in the valid set. For example,
// if the valid set is "1234567890", it will consume all digits in the number "123 "
// and will stop at the whitespace.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.readChar()) {
	}
	l.backup()
}

// switch2 is a helper function to choose between 2 available tokens based on
// the initial rune. You start with a first byte
// e.g. '+' or '=' and then you check if the next byte is '='. This one is useful
// for comparison and assignment operators.
// The switch helpers are based on the switches implemented in the official GO lexer.
func (l *lexer) switch2(tkn0, tkn1 token.TokenType) token.TokenType {
	if l.accept("=") {
		return tkn1
	}
	return tkn0
}

// switch3 is a helper function to choose between 3 available tokens based
// on the initial rune.
func (l *lexer) switch3(
	tkn0, tkn1 token.TokenType,
	char string, tkn2 token.TokenType) token.TokenType {
	if l.accept("=") {
		return tkn1
	}

	if l.accept(char) {
		return tkn2
	}
	return tkn0
}

/* switch4 is a helper function to choose between 4 available tokens based
* on the initial rune.
* In the following example we start with '<' and then we check the next byte.
* We can either stop at '<', go to '<=', or go to '<<'. If we go to '<<', we can
* go proceed to '<<='.
*              <
*            /  |  \
*           /   |   \
*          LT   =    <
*              /    /  \
*             LTE  /    =
*                 /      \
*                SHL    ASSIGN_SHL
* */
func (l *lexer) switch4(
	tkn0, tkn1 token.TokenType, char string,
	tkn2, tkn3 token.TokenType) token.TokenType {
	if l.accept("=") {
		return tkn1
	}

	if l.accept(char) {
		if l.accept("=") {
			return tkn3
		}
		return tkn2
	}

	return tkn0
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
