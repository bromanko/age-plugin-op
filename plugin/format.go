package plugin

import (
	"bytes"
	"encoding/base64"
	"io"
)

var b64 = base64.RawStdEncoding.Strict()

const ColumnsPerLine = 64

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
