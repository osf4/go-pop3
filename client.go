package pop3

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/benpate/convert"
	"github.com/emersion/go-message"
)

// MessageInfo contains message ID with size or UIDL
type MessageInfo struct {
	ID int

	Size int
	UIDL string
}

// Client represents a POP3 client
type Client struct {
	server net.Conn

	rd *bufio.Scanner
	wr *bufio.Writer
}

// UidlAll returns a list of pairs(msg.ID and msg.UIDL)
func (c *Client) UidlAll() (msgs []MessageInfo, err error) {
	res, err := c.cmd(Uidl, true)
	if err != nil {
		return nil, err
	}

	line := 2
	msgsInfo := res.LinesFrom(line) // first line is not a pair of msg.ID and msg.Uidl

	msgs = make([]MessageInfo, len(msgsInfo))

	for i := range msgsInfo {
		args := res.Args(line)

		msgs[i].ID = convert.Int(args[0])
		msgs[i].UIDL = args[1]

		line++
	}

	return
}

// Uidl returns the Unique ID of the message
func (c *Client) Uidl(n int) (msg *MessageInfo, err error) {
	res, err := c.cmd(Uidl, false, n)
	if err != nil {
		return nil, err
	}

	msg = new(MessageInfo)
	args := res.Args(1)

	msg.ID = convert.Int(args[0])
	msg.UIDL = args[1]

	return
}

// Top returns first n lines of the message as *message.Entity
//
// See more at emersion/go-message
func (c *Client) Top(id, n int) (msg *message.Entity, err error) {
	res, err := c.cmd(TOP, true, id, n)
	if err != nil {
		return nil, err
	}

	rd := strings.NewReader(res.JoinFrom(2)) // message starts from second line
	return message.Read(rd)
}

// TopRaw returns first n lines of the message
func (c *Client) TopRaw(id, n int) (lines []string, err error) {
	res, err := c.cmd(TOP, true, id, n)
	if err != nil {
		return nil, err
	}

	return res.LinesFrom(2), nil
}

// Rset unmarks the messages marked for deletion
func (c *Client) Rset() (err error) {
	_, err = c.cmd(RSET, false)
	return err
}

// Dele removes a message from the mailbox
func (c *Client) Dele(n int) (err error) {
	_, err = c.cmd(DELE, false, n)
	return err
}

// Retr downloads message N and returns it as *message.Entity
//
// See more at emersion/go-message
func (c *Client) Retr(n int) (msg *message.Entity, err error) {
	res, err := c.cmd(RETR, true, n)
	if err != nil {
		return nil, err
	}

	rd := strings.NewReader(res.JoinFrom(2))
	return message.Read(rd)
}

// RetrRaw returns the size and the raw body of the message
func (c *Client) RetrRaw(n int) (text string, err error) {
	res, err := c.cmd(RETR, true, n)
	if err != nil {
		return "", err
	}

	return res.JoinFrom(2), nil
}

// ListAll returns a list of pairs(msg.ID and msg.Size)
func (c *Client) ListAll() (msgs []MessageInfo, err error) {
	res, err := c.cmd(LIST, true)
	if err != nil {
		return nil, err
	}

	line := 2
	msgsInfo := res.LinesFrom(line) // first line is not a pair of msg.ID and msg.Size

	msgs = make([]MessageInfo, len(msgsInfo))

	for i := range msgsInfo {
		args := res.Args(line)

		msgs[i].ID = convert.Int(args[0])
		msgs[i].Size = convert.Int(args[1])

		line++
	}

	return
}

// List returns the size of the message
func (c *Client) List(n int) (msg *MessageInfo, err error) {
	res, err := c.cmd(LIST, false, n)
	if err != nil {
		return nil, err
	}

	msg = new(MessageInfo)
	args := res.Args(1)

	msg.ID = convert.Int(args[0])
	msg.Size = convert.Int(args[1])

	return
}

// Stat returns the number of messages and their total size in bytes in the inbox.
func (c *Client) Stat() (count, size int, err error) {
	res, err := c.cmd(STAT, false)
	if err != nil {
		return 0, 0, err
	}

	args := res.Args(1)
	count = convert.Int(args[0])
	size = convert.Int(args[1])

	return
}

// Auth sends the username and the password to the server
func (c *Client) Auth(user, pass string) (err error) {
	err = c.User(user)
	if err != nil {
		return err
	}

	return c.Pass(pass)
}

// User sends the username to the server
func (c *Client) User(user string) (err error) {
	_, err = c.cmd(USER, false, user)
	return err
}

// Pass sends the password to the server
func (c *Client) Pass(pass string) (err error) {
	_, err = c.cmd(PASS, false, pass)
	return err
}

// Noop does nothing and returns positive response
//
// It is used only to check the server connection
func (c *Client) Noop() (err error) {
	_, err = c.cmd(NOOP, false)
	return err
}

// Quit sends the quit command to the server
func (c *Client) Quit() {
	c.cmd(QUIT, false)
}

// Close sends the quit command to the server and closes the connection
func (c *Client) Close() error {
	c.Quit()
	return c.server.Close()
}

// cmd sends request to the server and reads response
func (c *Client) cmd(cmd command, multiline bool, args ...any) (res *Response, err error) {
	err = c.sendRequest(cmd, args...)
	if err != nil {
		return nil, err
	}

	return c.readResponse(multiline)
}

func (c *Client) sendRequest(cmd command, args ...any) (err error) {
	req := NewRequest(cmd, args...)

	_, err = io.WriteString(c.wr, req.String())
	if err != nil {
		return err
	}

	return c.wr.Flush()
}

func (c *Client) readResponse(multiline bool) (res *Response, err error) {
	var lines []string

	if multiline {
		lines, err = c.scanLines()
	} else {
		lines = make([]string, 1)
		lines[0], _, err = c.scan()
	}

	if err != nil {
		return nil, err
	}

	res = new(Response)

	err = res.Parse(lines)
	if err != nil {
		return nil, err
	}

	// If status code is '-ERR' the first line contains error message
	if res.Code == ERR {
		return nil, fmt.Errorf("pop3: %v", res.Text[0])
	}

	return
}

// scanLines reads lines of the response, until the line is a termination octet
func (c *Client) scanLines() (lines []string, err error) {
	var (
		line string
		eof  bool
	)

	for !eof {
		line, eof, err = c.scan()
		if err != nil {
			break
		}

		lines = append(lines, line)
	}

	return
}

// scan returns a single line of response
//
// eof == true if the line is a termination octet
func (c *Client) scan() (line string, eof bool, err error) {
	ok := c.rd.Scan()
	if !ok {
		return "", true, c.rd.Err()
	}

	line = c.rd.Text()

	eof = line == "."
	return line, eof, nil
}
