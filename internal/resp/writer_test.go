package resp

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		name  string
		write func(w *Writer)
		want  string
	}{
		{"simple string", func(w *Writer) { w.WriteSimpleString("OK") }, "+OK\r\n"},
		{"erro", func(w *Writer) { w.WriteError("ERR falhou") }, "-ERR falhou\r\n"},
		{"inteiro", func(w *Writer) { w.WriteInteger(42) }, ":42\r\n"},
		{"inteiro negativo", func(w *Writer) { w.WriteInteger(-7) }, ":-7\r\n"},
		{"bulk string", func(w *Writer) { w.WriteBulkString("hello") }, "$5\r\nhello\r\n"},
		{"bulk string vazia", func(w *Writer) { w.WriteBulkString("") }, "$0\r\n\r\n"},
		{"bulk com acento", func(w *Writer) { w.WriteBulkString("não") }, "$4\r\nnão\r\n"},
		{"nulo", func(w *Writer) { w.WriteNull() }, "$-1\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewWriter(&buf)
			tt.write(w)
			if err := w.Flush(); err != nil {
				t.Fatalf("erro no flush: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriter_SemFlushNadaEhEnviado(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.WriteSimpleString("OK")

	if buf.Len() != 0 {
		t.Fatalf("esperava buffer vazio antes do Flush, got %q", buf.String())
	}

	w.Flush()
	if buf.String() != "+OK\r\n" {
		t.Errorf("got %q depois do Flush", buf.String())
	}
}
