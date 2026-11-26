package resp

import (
	"bufio"
	"io"
	"strconv"
)

type Writer struct {
	wr *bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{wr: bufio.NewWriter(w)}
}

func (w *Writer) WriteSimpleString(s string) {
	w.wr.WriteByte(byte(SimpleString))
	w.wr.WriteString(s)
	w.wr.WriteString("\r\n")
}

func (w *Writer) WriteError(msg string) {
	w.wr.WriteByte(byte(Error))
	w.wr.WriteString(msg)
	w.wr.WriteString("\r\n")
}

func (w *Writer) WriteInteger(n int64) {
	w.wr.WriteByte(byte(Integer))
	w.wr.WriteString(strconv.FormatInt(n, 10))
	w.wr.WriteString("\r\n")
}

func (w *Writer) WriteBulkString(s string) {
	w.wr.WriteByte(byte(BulkString))
	w.wr.WriteString(strconv.Itoa(len(s)))
	w.wr.WriteString("\r\n")
	w.wr.WriteString(s)
	w.wr.WriteString("\r\n")
}

func (w *Writer) WriteNull() {
	w.wr.WriteString("$-1\r\n")
}

func (w *Writer) Flush() error {
	return w.wr.Flush()
}
