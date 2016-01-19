package stream

import (
	"crypto/tls"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
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

type Stream struct {
	URL     *url.URL
	prefix  []byte
	Input   io.Reader
	Output  io.Writer
	CloseCh chan bool
}

func NewStream(uri string, pref string, prefColor Color) (*Stream, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	s := &Stream{URL: u}

	path := u.Host + u.Path

	if u.Scheme == "" && path != "" {
		u.Scheme = "file"
	}

	switch u.Scheme {
	case "":
		s.Input = os.Stdin //use standard input, output
		s.Output = os.Stdout
	case "file":
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		s.Output = f //if stdio is a file, do not support stdin (not intercative)

	case "ssl":
		fallthrough
	case "tls":
		tcpConn, err := net.DialTimeout("tcp", path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{InsecureSkipVerify: true}
		conn := tls.Client(tcpConn, config)
		s.Input = conn
		s.Output = conn

	default:
		conn, err := net.DialTimeout(u.Scheme, path, DIAL_TIMEOUT)
		if err != nil {
			return nil, err
		}
		s.Input = conn
		s.Output = conn
	}

	if pref != "" {
		if prefColor == NoColor {
			s.prefix = []byte(pref)
		} else {
			s.prefix = []byte(escapeCode(prefColor) + pref + resetEscapeCode())
		}
	}

	s.CloseCh = make(chan bool, 10)

	return s, nil
}

func (s *Stream) Write(p []byte) (int, error) {
	if len(s.prefix) == 0 {
		return s.Output.Write(p)
	}

	if strings.Trim(string(p), "\n\t") == "" {
		return s.Output.Write(p)
	}

	n, err := s.Output.Write(append(s.prefix, p...))

	//the caller will except the write to be len(p), so if we have a prefix, it will confused it.
	//That is why we need to check for an ErrShortWrite error and return n = n - len(prefix)

	if err != nil && err != io.ErrShortWrite {
		return n, err
	}
	return n - len(s.prefix), err
}

func (s *Stream) Read(p []byte) (int, error) {
	if s.Input == nil {
		return 0, nil //if file, no write
	}
	return s.Input.Read(p)
}

func (s *Stream) Close() {
	if rc, ok := s.Input.(io.ReadCloser); ok && s.Input != os.Stdin {
		rc.Close()
	}

	if wc, ok := s.Output.(io.WriteCloser); ok && s.Output != os.Stdout {
		wc.Close()
	}
	s.CloseCh <- true
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
