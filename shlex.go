package shlex

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

var ErrInvalid = errors.New("invalid word")

type runeWriter interface {
	WriteRune(rune) (int, error)
}

func SplitString(str string) ([]string, error) {
	return Split(strings.NewReader(str))
}

func Split(r io.Reader) ([]string, error) {
	return split(bufio.NewReader(r))
}

func split(rs *bufio.Reader) ([]string, error) {
	var (
		buf bytes.Buffer
		str []string
	)
	for {
		r, _, err := rs.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		switch {
		case isNL(r) || isBlank(r):
			readBlank(rs)
			continue
		case isBrace(r):
			err = readBrace(&buf, rs)
		case isParen(r):
			err = readGroup(&buf, rs)
		case isDollar(r):
			err = readDollar(&buf, rs)
		case isQuote(r):
			err = readQuote(&buf, rs, r)
		case isDelimiter(r):
			readDelimiter(&buf, rs, r)
		case isComment(r):
			readComment(&buf, rs)
		default:
			readWord(&buf, rs, r)
		}
		if err != nil {
			return str, err
		}
		str = append(str, buf.String())
		buf.Reset()
	}
	return str, nil
}

func readComment(str runeWriter, rs io.RuneScanner) error {
	str.WriteRune(dash)
	for {
		r, _, err := rs.ReadRune()
		if err != nil {
			break
		}
		str.WriteRune(r)
	}
	return nil
}

func readDollar(str runeWriter, rs io.RuneScanner) error {
	if r, _, _ := rs.ReadRune(); r != lparen {
		rs.UnreadRune()
		readWord(str, rs, dollar)
		return nil
	}
	if r, _, _ := rs.ReadRune(); r == lparen {
		return readArithmetic(str, rs)
	}
	rs.UnreadRune()
	return readSubstitution(str, rs)
}

func readSubstitution(str runeWriter, rs io.RuneScanner) error {
	str.WriteRune(dollar)
	str.WriteRune(lparen)

	for {
		r, _, err := rs.ReadRune()
		if err != nil {
			return ErrInvalid
		}
		if r == dollar {
			if err = readDollar(str, rs); err != nil {
				return err
			}
			continue
		}
		if r == rparen {
			break
		}
		str.WriteRune(r)
	}
	str.WriteRune(rparen)
	return nil
}

func readArithmetic(str runeWriter, rs io.RuneScanner) error {
	str.WriteRune(dollar)
	str.WriteRune(lparen)
	str.WriteRune(lparen)

	var prev rune
	for {
		r, _, err := rs.ReadRune()
		if err != nil {
			return ErrInvalid
		}
		if r == dollar {
			if err = readDollar(str, rs); err != nil {
				return err
			}
			continue
		}
		if r == rparen && prev == rparen {
			break
		}
		if r == lparen {
			if err = readGroup(str, rs); err != nil {
				return err
			}
			continue
		}
		prev = r
		str.WriteRune(r)
	}
	str.WriteRune(rparen)
	return nil
}

func readBrace(str runeWriter, rs io.RuneScanner) error {
	str.WriteRune(lcurly)
	for {
		c, _, err := rs.ReadRune()
		if err != nil {
			return ErrInvalid
		}
		if c == rcurly {
			break
		}
		if c == lcurly {
			if err = readBrace(str, rs); err != nil {
				return err
			}
			continue
		}
		str.WriteRune(c)
	}
	str.WriteRune(rcurly)
	return nil
}

func readGroup(str runeWriter, rs io.RuneScanner) error {
	str.WriteRune(lparen)
	for {
		c, _, err := rs.ReadRune()
		if err != nil {
			return ErrInvalid
		}
		if c == rparen {
			break
		}
		if c == lparen {
			if err = readGroup(str, rs); err != nil {
				return err
			}
			continue
		}
		str.WriteRune(c)
	}
	str.WriteRune(rparen)
	return nil
}

func readWord(str runeWriter, rs io.RuneScanner, r rune) {
	str.WriteRune(r)
	var unread bool
	for {
		r, _, err := rs.ReadRune()
		if eow(r) || err != nil {
			unread = r != equal
			break
		}
		str.WriteRune(r)
	}
	if unread {
		rs.UnreadRune()
	}
}

func readDelimiter(str runeWriter, rs io.RuneScanner, r rune) {
	str.WriteRune(r)
	for {
		r, _, err := rs.ReadRune()
		if !isDelimiter(r) || err != nil {
			break
		}
		str.WriteRune(r)
	}
	rs.UnreadRune()
}

func readQuote(str runeWriter, rs io.RuneReader, quote rune) error {
	var prev rune
	for {
		r, _, err := rs.ReadRune()
		if err != nil {
			return ErrInvalid
		}
		if r == quote && prev != backslash {
			break
		}
		prev = r
		str.WriteRune(r)
	}
	return nil
}

func readBlank(rs io.RuneScanner) {
	for {
		r, _, _ := rs.ReadRune()
		if !isNL(r) && !isBlank(r) {
			break
		}
	}
	rs.UnreadRune()
}

const (
	ampersand = '&'
	pipe      = '|'
	semicolon = ';'
	space     = ' '
	tab       = '\t'
	squote    = '\''
	dquote    = '"'
	backslash = '\\'
	nl        = '\n'
	cr        = '\r'
	dollar    = '$'
	lparen    = '('
	rparen    = ')'
	lcurly    = '{'
	rcurly    = '}'
	dash      = '#'
	equal     = '='
	minus     = '-'
)

func eow(r rune) bool {
	return isDelimiter(r) || isQuote(r) || isBlank(r) || isNL(r) || r == equal
}

func isParen(r rune) bool {
	return r == lparen
}

func isBrace(r rune) bool {
	return r == lcurly
}

func isComment(r rune) bool {
	return r == dash
}

func isDollar(r rune) bool {
	return r == dollar
}

func isDelimiter(r rune) bool {
	return r == ampersand || r == pipe || r == semicolon
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isDouble(r rune) bool {
	return r == dquote
}

func isSingle(r rune) bool {
	return r == squote
}

func isQuote(r rune) bool {
	return isDouble(r) || isSingle(r)
}

func isNL(r rune) bool {
	return r == cr || r == nl
}
