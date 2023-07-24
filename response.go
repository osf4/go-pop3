package pop3

import (
	"errors"
	"strings"
)

type code string

const (
	OK  code = "+OK"
	ERR code = "-ERR"
)

var (
	ErrUnknownCode = errors.New("pop3: unknown response code")
)

// Response represents replies sent by the server
type Response struct {
	Code code
	Text []string // Text contains lines of the response
}

// Args returns arguments on the specified line
func (r *Response) Args(n int) []string {
	if n <= 0 || n > len(r.Text) {
		n = 1
	}

	return strings.Fields(r.Text[n-1])
}

// LinesFrom returns lines of the response starting from n
func (r *Response) LinesFrom(n int) []string {
	if n <= 0 || n > len(r.Text) {
		n = 1
	}

	return r.Text[n-1:]
}

// JoinLines join lines of the response into one string
//
// Each line is separated by the CRLF pair
func (r *Response) JoinFrom(from int) string {
	if from <= 0 || from > len(r.Text) {
		from = 1
	}

	return strings.Join(r.Text[from-1:len(r.Text)], CRLF)
}

func (r *Response) Multiline() bool {
	return len(r.Text) > 1
}

func (r *Response) Parse(lines []string) (err error) {
	c, text, withText := strings.Cut(lines[0], " ")
	if !isStatusCode(c) {
		return ErrUnknownCode
	}

	r.Code = code(c)
	if withText {
		r.Text = lines
		r.Text[0] = text

		if r.Multiline() {
			return r.parseMultiline()
		}
	}

	return nil
}

// parseMultiline removes the line containing a termination octet
//
// Returns error if the last line of response is not '.'
func (r *Response) parseMultiline() error {
	if r.Text[len(r.Text)-1] != "." {
		return errors.New("pop3: termination octet in multiline response is missing")
	}

	r.Text = r.Text[:len(r.Text)-1]

	return nil
}

// isStatusCode returns false if s is not '+OK' or '-ERR'
func isStatusCode(s string) bool {
	switch c := code(s); c {
	case OK, ERR:
		return true

	default:
		return false
	}
}
