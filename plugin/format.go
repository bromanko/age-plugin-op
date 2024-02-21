package plugin

// Copyright 2019 The age Authors: https://github.com/FiloSottile/age/blob/29b68c20fc241bf2e11bdd3e59b4368fe689e12a/AUTHORS
// Copied from https://github.com/FiloSottile/age/blob/29b68c20fc241bf2e11bdd3e59b4368fe689e12a/internal/format/format.go#L295

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"filippo.io/age"
	"fmt"
	"io"
	"strings"
)

var b64 = base64.RawStdEncoding.Strict()

func DecodeString(s string) ([]byte, error) {
	// CR and LF are ignored by DecodeString, but we don't want any malleability.
	if strings.ContainsAny(s, "\n\r") {
		return nil, errors.New(`unexpected newline character`)
	}
	return b64.DecodeString(s)
}

const ColumnsPerLine = 64

const BytesPerLine = ColumnsPerLine / 4 * 3

// NewWrappedBase64Encoder returns a WrappedBase64Encoder that writes to dst.
func NewWrappedBase64Encoder(enc *base64.Encoding, dst io.Writer) *WrappedBase64Encoder {
	w := &WrappedBase64Encoder{dst: dst}
	w.enc = base64.NewEncoder(enc, WriterFunc(w.writeWrapped))
	return w
}

type WriterFunc func(p []byte) (int, error)

func (f WriterFunc) Write(p []byte) (int, error) { return f(p) }

// WrappedBase64Encoder is a standard base64 encoder that inserts an LF
// character every ColumnsPerLine bytes. It does not insert a newline neither at
// the beginning nor at the end of the stream, but it ensures the last line is
// shorter than ColumnsPerLine, which means it might be empty.
type WrappedBase64Encoder struct {
	enc     io.WriteCloser
	dst     io.Writer
	written int
	buf     bytes.Buffer
}

func (w *WrappedBase64Encoder) Write(p []byte) (int, error) { return w.enc.Write(p) }

func (w *WrappedBase64Encoder) Close() error {
	return w.enc.Close()
}

func (w *WrappedBase64Encoder) writeWrapped(p []byte) (int, error) {
	if w.buf.Len() != 0 {
		panic("age: internal error: non-empty WrappedBase64Encoder.buf")
	}
	for len(p) > 0 {
		toWrite := ColumnsPerLine - (w.written % ColumnsPerLine)
		if toWrite > len(p) {
			toWrite = len(p)
		}
		n, _ := w.buf.Write(p[:toWrite])
		w.written += n
		p = p[n:]
		if w.written%ColumnsPerLine == 0 {
			w.buf.Write([]byte("\n"))
		}
	}
	if _, err := w.buf.WriteTo(w.dst); err != nil {
		// We always return n = 0 on error because it's hard to work back to the
		// input length that ended up written out. Not ideal, but Write errors
		// are not recoverable anyway.
		return 0, err
	}
	return len(p), nil
}

var stanzaPrefix = []byte("->")
var footerPrefix = []byte("---")

type StanzaReader struct {
	r   *bufio.Reader
	err error
}

func NewStanzaReader(r *bufio.Reader) *StanzaReader {
	return &StanzaReader{r: r}
}

func (r *StanzaReader) ReadStanza() (s *age.Stanza, err error) {
	// Read errors are unrecoverable.
	if r.err != nil {
		return nil, r.err
	}
	defer func() { r.err = err }()

	s = &age.Stanza{}

	line, err := r.r.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}
	if !bytes.HasPrefix(line, stanzaPrefix) {
		return nil, fmt.Errorf("malformed stanza opening line: %q", line)
	}
	prefix, args := SplitArgs(line)
	if prefix != string(stanzaPrefix) || len(args) < 1 {
		return nil, fmt.Errorf("malformed stanza: %q", line)
	}
	for _, a := range args {
		if !isValidString(a) {
			return nil, fmt.Errorf("malformed stanza: %q", line)
		}
	}
	s.Type = args[0]
	s.Args = args[1:]

	for {
		line, err := r.r.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}

		b, err := DecodeString(strings.TrimSuffix(string(line), "\n"))
		if err != nil {
			if bytes.HasPrefix(line, footerPrefix) || bytes.HasPrefix(line, stanzaPrefix) {
				return nil, fmt.Errorf("malformed body line %q: stanza ended without a short line\nNote: this might be a file encrypted with an old beta version of age or rage. Use age v1.0.0-beta6 or rage to decrypt it.", line)
			}
			return nil, errorf("malformed body line %q: %v", line, err)
		}
		if len(b) > BytesPerLine {
			return nil, errorf("malformed body line %q: too long", line)
		}
		s.Body = append(s.Body, b...)
		if len(b) < BytesPerLine {
			// A stanza body always ends with a short line.
			return s, nil
		}
	}
}

type ParseError struct {
	err error
}

func (e *ParseError) Error() string {
	return "parsing age header: " + e.err.Error()
}

func (e *ParseError) Unwrap() error {
	return e.err
}

func errorf(format string, a ...interface{}) error {
	return &ParseError{fmt.Errorf(format, a...)}
}

func SplitArgs(line []byte) (string, []string) {
	l := strings.TrimSuffix(string(line), "\n")
	parts := strings.Split(l, " ")
	return parts[0], parts[1:]
}

func isValidString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}
