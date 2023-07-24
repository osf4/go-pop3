package pop3

import (
	"bufio"
	"crypto/tls"
	"net"
	"time"
)

const (
	CRLF = "\r\n"
)

// Opt represents the client configuration
type Opt struct {
	DialTimeout time.Duration

	TLSEnabled    bool
	TLSSkipVerify bool
}

var (
	defaultOpt = &Opt{
		DialTimeout: time.Second * 3,
		TLSEnabled:  true,
	}
)

// Dial returns POP3 connection to the server
func Dial(addr string, opt *Opt) (*Client, error) {
	if opt == nil {
		opt = defaultOpt
	}

	if opt.DialTimeout < time.Second {
		opt.DialTimeout = time.Second * 3
	}

	server, err := net.DialTimeout("tcp", addr, opt.DialTimeout)
	if err != nil {
		return nil, err
	}

	if opt.TLSEnabled {
		// Need hostname to verify the server
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		config := new(tls.Config)
		if opt.TLSSkipVerify {
			config.InsecureSkipVerify = true
		} else {
			config.ServerName = host
		}

		server = tls.Client(server, config)
	}

	client := &Client{
		server: server,
		wr:     bufio.NewWriter(server),
		rd:     bufio.NewScanner(server),
	}

	// Verify the connection by reading the welcome greeting
	_, err = client.readResponse(false)
	if err != nil {
		return nil, err
	}

	return client, nil
}
