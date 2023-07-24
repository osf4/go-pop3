package pop3

import (
	"strings"

	"github.com/benpate/convert"
)

type command string

const (
	QUIT command = "QUIT"
	STAT command = "STAT"
	LIST command = "LIST"
	RETR command = "RETR"
	DELE command = "DELE"
	NOOP command = "NOOP"
	RSET command = "RSET"
	TOP  command = "TOP"
	Uidl command = "UIDL"
	USER command = "USER"
	PASS command = "PASS"
	APOP command = "APOP"
)

// Request represents requests sent by client
type Request struct {
	Cmd  command
	Args []string
}

func NewRequest(cmd command, args ...any) *Request {
	return &Request{
		Cmd:  cmd,
		Args: convert.SliceOfString(args),
	}
}

// String returns string presentation of the request
func (r *Request) String() string {
	var s strings.Builder

	s.WriteString(string(r.Cmd))

	for _, arg := range r.Args {
		s.WriteRune(' ')
		s.WriteString(arg)
	}

	s.WriteString(CRLF)
	return s.String()
}
