package pop3

import (
	"fmt"
	"io"
	"net/smtp"
	"testing"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/stretchr/testify/assert"
)

const (
	MessagesToSend = 5
)

var (
	Message = "POP3 test message\r\nsecond line of the message"
)

func addMessages(n int) (err error) {
	from := "unknown@example.com"
	to := []string{
		"recipient@example.com",
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";"
	format := "From: %v\r\nTo: %v\r\nSubject: POP3 test mail(%v)\r\n%v\n\n%v\r\n"

	var msg string
	for i := 0; i < n; i++ {
		msg = fmt.Sprintf(format, from, to[0], i, mime, Message)

		err = smtp.SendMail("localhost:2500", nil, from, to, []byte(msg))
		if err != nil {
			break
		}
	}

	return err
}

func dialLocalhost() (*Client, error) {
	opt := &Opt{
		DialTimeout: time.Second * 3,
		TLSEnabled:  false,
	}

	return Dial("localhost:1100", opt)
}

func compareMessage(t *testing.T, m *message.Entity, msg string) {
	mr := mail.NewReader(m)

	if mr == nil {
		contentType, _, _ := m.Header.ContentType()
		t.Logf("This is a non-multipart message with type: %v\n", contentType)
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Errorf("error reading next part of the message: %v", err)
		}

		b, err := io.ReadAll(p.Body)
		if err != nil {
			t.Errorf("error reading body of the part: %v", err)
		}

		assert.Equal(t, msg, string(b))
	}
}

func TestAll(t *testing.T) {
	c, err := dialLocalhost()
	if err != nil {
		t.Fatalf("unable to establish connection to the POP server: %v", err)
	}

	err = addMessages(MessagesToSend)
	if err != nil {
		t.Fatalf("unable to send messages to the mail server: %v", err)
	}

	// testing Auth
	err = c.Auth("recipient", "password")
	if err != nil {
		t.Fatalf("unable to authenticate on the server: %v", err)
	}

	// testing Stat
	count, _, err := c.Stat()
	if err != nil {
		t.Fatalf("error using Stat: %v", err)
	}

	assert.Equal(t, MessagesToSend, count, "mailbox must contain 5 messages")

	// testing Uidl
	msgInfo, err := c.Uidl(1)
	if err != nil {
		t.Fatalf("error using Uidl: %v", err)
	}

	assert.Equal(t, 1, msgInfo.ID, "message id must be 1")

	// testing UidlAll
	msgsInfo, err := c.UidlAll()
	if err != nil {
		t.Fatalf("error using UidlAll: %v", err)
	}

	assert.Equal(t, MessagesToSend, len(msgsInfo), "UidlAll must return all messages from the mailbox")

	// testing List
	msgInfo, err = c.List(1)
	if err != nil {
		t.Fatalf("error using List: %v", err)
	}

	assert.Equal(t, 1, msgInfo.ID, "message id must be 1")

	msgsInfo, err = c.ListAll()
	if err != nil {
		t.Fatalf("error using ListAll: %v", err)
	}

	assert.Equal(t, MessagesToSend, len(msgsInfo), "mailbox must contain 5 messages")

	// testing Retr
	m, err := c.Retr(1)
	if err != nil {
		t.Fatalf("error using Retr: %v", err)
	}

	assert.Equal(t, "POP3 test mail(0)", m.Header.Get("Subject"), "Retr returned wrong subject")
	compareMessage(t, m, Message)

	// testing Top
	m, err = c.Top(1, 1)
	if err != nil {
		t.Fatalf("error using Top: %v", err)
	}

	compareMessage(t, m, "POP3 test message")

	// testing Noop
	err = c.Noop()
	if err != nil {
		t.Fatalf("error in using Noop: %v", err)
	}

	// testing Dele
	err = c.Dele(1)
	if err != nil {
		t.Fatalf("error using Dele: %v", err)
	}

	count, _, _ = c.Stat()

	assert.Equal(t, 4, count, "mailbox must contain only 4 messages after Dele command")

	// testing Rset, list
	err = c.Rset()
	if err != nil {
		t.Fatalf("error using Rset: %v", err)
	}

	count, _, _ = c.Stat()
	assert.Equal(t, 5, count, "mailbox must contain 5 messages after using Rset")
}
