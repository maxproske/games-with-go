package apt

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Enum
type tokenType int

const (
	openParen tokenType = iota // First is 1, next 2, ...
	closeParen
	op
	constant // We have to parse these as numbers, so keep them seperate
)

type token struct {
	typ   tokenType // Is it an open paren, number, etc.
	value string    // Name of operator we found
}

// Keeps track of where we are in the text
type lexer struct {
	input  string     // Entirety of file we are loading in
	start  int        // Points to start of whatever token we are working on
	pos    int        // Actual char we are pointed at in the string
	width  int        // Width of the current character (unicode support)
	tokens chan token // Have lexer running as it gets tokens!
}

// Break text into list of things we're interested in (lexing). Ignore tabs, carriage returns, etc.
// stateFunction takes a pointer to a lexer, returns another state function
type stateFunc func(*lexer) stateFunc

func parse(tokens chan token) Node {
	for {
		token, ok := <-tokens
		if !ok {
			// Check if token channel has been closed
			panic("No more tokens")
		}
		fmt.Print(token.value, ",")
	}
	return nil // Just lex for now, no parsing
}

const eof rune = -1

// Returns a node to kick off the parser at the same time
func BeginLexing(s string) Node {
	// Make a buffered channel (100). Lexer can keep putting tokens into the channel, even if the parser hasn't pulled them out yet.
	// (Lexer can keep moving if the parser is lagging behind).
	l := &lexer{input: s, tokens: make(chan token, 100)} // name:value format = we don't have to provide in order.

	go l.run() // Run lexer on a seperate thread

	return parse(l.tokens) // Parser takes tokens
}

func (l *lexer) run() {
	// Magic loop to go from one state to another. Set state equal to some type stateFunc.
	// First statefunction that gets called is determineToken
	for state := determineToken; state != nil; {
		state = state(l) // For each loop, state will equal the state it gets back. (for isStartOfNumber)
	}
	close(l.tokens) // Close the channel
}

// State function, start looking for first token to grab
func determineToken(l *lexer) stateFunc {
	for {
		// Grab the next rune (char) in the string. Decide what to do with it
		switch r := l.next(); {
		case isWhiteSpace(r):
			l.ignore() // Ignore it
		case r == '(':
			l.emit(openParen) // Ready to spit out a token
		case r == ')':
			l.emit(closeParen)
		case isStartOfNumber(r):
			return lexNumber // Parse a number. Return statefunc that needs to run next
		case r == eof:
			return nil // Close program
		default:
			// has to be an Op
			return lexOp
		}
	}
}

func lexNumber(l *lexer) stateFunc {
	l.accept("+-.")
	digits := "0123456789"
	l.acceptRun(digits)
	// Once we have run out of digits to accept, see if we can accept a period
	if l.accept(".") {
		l.acceptRun(digits) // 123.123
	}
	// Account for OpMinus (if all we got was a minus)
	if l.input[l.start:l.pos] == "-" {
		l.emit(op)
	} else {
		l.emit(constant)
	}

	return determineToken // Go back to determine token state
}

func lexOp(l *lexer) stateFunc {
	l.acceptRun("+-/*abcdefhijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123546789") // Match these
	l.emit(op)
	return determineToken

}

// lexNumber helper function (1 char)
func (l *lexer) accept(valid string) bool {
	// Is the next input one of these things or not
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// accept() for multiple chars
func (l *lexer) acceptRun(valid string) {
	// Keep advancing as long as it exists in the string of valid characters
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	// Then backup
	l.backup()
}

func isWhiteSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

func isStartOfNumber(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == '+' || r == '.'
}

// Rune is a go type that is a "more than a byte" character (unicode)
func (l *lexer) next() (r rune) {

	if l.pos >= len(l.input) {
		l.width = 0
		return eof // We're done with the whole file
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:]) // Get slice from current position of lexor to the end, and width of the next rune
	l.pos += l.width                                      // Increase position with width of current char
	return r                                              // Return rune
}

// If next character is not what we expect, backup
func (l *lexer) backup() {
	l.pos -= l.width
}

// Go further in the input while ignoring (eg. whitespace)
func (l *lexer) ignore() {
	l.start = l.pos // Not considering start as part of the token. Just skip it.
}

// Look at current character without advancing the lexer
func (l *lexer) peek() (r rune) {
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:]) // Same decode without updating position
	return r
}

// Tell lexer to go ahead and spit out a token
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]} // Write to tokens channel from start to pos
	l.start = l.pos                              // Move on to next token
}
