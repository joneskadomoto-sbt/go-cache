package resp

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReader_Read(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{"simple string", "+OK\r\n", Value{Type: SimpleString, Str: "OK"}},
		{"erro", "-ERR algo deu errado\r\n", Value{Type: Error, Str: "ERR algo deu errado"}},
		{"inteiro", ":1000\r\n", Value{Type: Integer, Int: 1000}},
		{"inteiro negativo", ":-42\r\n", Value{Type: Integer, Int: -42}},
		{"bulk string", "$5\r\nhello\r\n", Value{Type: BulkString, Str: "hello"}},
		{"bulk string vazia", "$0\r\n\r\n", Value{Type: BulkString, Str: ""}},
		{"bulk com \\r\\n dentro", "$4\r\na\r\nb\r\n", Value{Type: BulkString, Str: "a\r\nb"}},
		{"bulk nula", "$-1\r\n", Value{Type: BulkString, Null: true}},
		{"array vazio", "*0\r\n", Value{Type: Array, Array: []Value{}}},
		{"array nulo", "*-1\r\n", Value{Type: Array, Null: true}},
		{
			"comando SET",
			"*3\r\n$3\r\nSET\r\n$4\r\nnome\r\n$5\r\nJones\r\n",
			Value{Type: Array, Array: []Value{
				{Type: BulkString, Str: "SET"},
				{Type: BulkString, Str: "nome"},
				{Type: BulkString, Str: "Jones"},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			got, err := r.Read()
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestReader_ReadErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"tipo desconhecido", "?abc\r\n", ErrProtocol},
		{"linha sem \\r", "+OK\n", ErrProtocol},
		{"inteiro inválido", ":abc\r\n", ErrProtocol},
		{"bulk com tamanho negativo", "$-5\r\n", ErrProtocol},
		{"bulk sem terminador", "$3\r\nabcXY", ErrProtocol},
		{"bulk incompleta", "$10\r\nabc", io.ErrUnexpectedEOF},
		{"entrada vazia", "", io.EOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.input))
			_, err := r.Read()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got erro %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestReader_Pipeline(t *testing.T) {
	input := "*1\r\n$4\r\nPING\r\n*2\r\n$4\r\nECHO\r\n$2\r\noi\r\n"
	r := NewReader(strings.NewReader(input))

	want := [][]string{{"PING"}, {"ECHO", "oi"}}
	for i, w := range want {
		v, err := r.Read()
		if err != nil {
			t.Fatalf("comando %d: erro inesperado: %v", i, err)
		}
		args, err := v.Args()
		if err != nil {
			t.Fatalf("comando %d: Args: %v", i, err)
		}
		if !reflect.DeepEqual(args, w) {
			t.Errorf("comando %d: got %v, want %v", i, args, w)
		}
	}

	if _, err := r.Read(); !errors.Is(err, io.EOF) {
		t.Errorf("esperava io.EOF no fim, got %v", err)
	}
}
