package iowire

import (
	"crypto/tls"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/pkg/term"
)

const (
	DIAL_TIMEOUT = 5 * time.Second
)

type Color uint8

const (
	White = iota
	Green
	Blue
	Magenta
	Yellow
	Cyan
	Red
	NoColor
)

type Wire struct {
	URL     *url.URL
	prefix  []byte
	Input   io.Reader
	Output  io.Writer
	CloseCh chan bool
}

func NewWire(uri string) (*Wire, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	wire := &Wire{URL: u}

	path := u.Host + u.Path

	if u.Scheme == "" && path != "" {
		u.Scheme = "file"
	}

	switch u.Scheme {
	case "":
		wire.Input = os.Stdin //use standard input, output
		wire.Output = os.Stdout
	case "file":
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		wire.Output = f
		wire.Input = os.Stdin

	case "ssl":
		fallthrough
	case "tls":
		tcpConn, err := net.DialTimeout("tcp", path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{InsecureSkipVerify: true}
		conn := tls.Client(tcpConn, config)
		wire.Input = conn
		wire.Output = conn

	default:
		conn, err := net.DialTimeout(u.Scheme, path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		wire.Input = conn
		wire.Output = conn
	}

	wire.CloseCh = make(chan bool, 10)

	return wire, nil
}

func (wire *Wire) SetPrefix(prefix string, color Color) {
	if color == NoColor {
		wire.prefix = []byte(prefix)
	} else {
		wire.prefix = []byte(escapeCode(color) + prefix + resetEscapeCode())
	}
	wire.Write(wire.prefix) //flush first prefix
}

//tell whether or not the stream is interactive
func (wire *Wire) Interactive() bool {
	_, isConnIn := wire.Input.(net.Conn)
	_, isConnOut := wire.Output.(net.Conn)
	if isConnIn && isConnOut {
		return true //assume connection stream are interactive
	}

	return wire.Terminal()
}

func (wire *Wire) Terminal() bool {
	_, isTerminalIn := term.GetFdInfo(wire.Input)
	_, isTerminalOut := term.GetFdInfo(wire.Output)
	return isTerminalIn && isTerminalOut
}

func (wire *Wire) Write(p []byte) (int, error) {
	if len(wire.prefix) == 0 || !strings.HasSuffix(string(p), "\n") {
		return wire.Output.Write(p)
	}

	//will write a line, write the prefix for next line
	n, err := wire.Output.Write(append(p, wire.prefix...))

	//the caller will except the write to be len(p), so if we have a prefix, it will confused it.
	//That is why we need to check for an ErrShortWrite error and return n = n - len(prefix)

	if err != nil && err != io.ErrShortWrite {
		return n, err
	}
	return n - len(wire.prefix), err
}

func (wire *Wire) Read(p []byte) (int, error) {
	if wire.Input == nil {
		return 0, nil
	}
	return wire.Input.Read(p)
}

func (wire *Wire) Close() {
	for _, i := range []interface{}{wire.Input, wire.Output} {
		if c, ok := i.(io.ReadCloser); ok {
			c.Close()
		}
	}
	wire.CloseCh <- true
}

func MapColor(c string) Color {
	switch c {
	case "red":
		return Red
	case "green":
		return Green
	case "blue":
		return Blue
	case "yellow":
		return Yellow
	case "magenta":
		return Magenta
	case "cyan":
		return Cyan
	case "white":
		return White
	default:
		return NoColor
	}
}

func escapeCode(color Color) string {
	switch color {
	case Red:
		return "\x1b[31;1m"
	case Green:
		return "\x1b[32;1m"
	case Blue:
		return "\x1b[34;1m"
	case Yellow:
		return "\x1b[33;1m"
	case Magenta:
		return "\x1b[35;1m"
	case Cyan:
		return "\x1b[36;1m"
	case White:
		return "\x1b[37;1m"
	case NoColor:
		fallthrough
	default:
		return ""
	}
}

func resetEscapeCode() string {
	return "\x1b[0m"
}
