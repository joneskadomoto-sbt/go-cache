package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Type byte

const (
	SimpleString Type = '+'
	Error        Type = '-'
	Integer      Type = ':'
	BulkString   Type = '$'
	Array        Type = '*'
)

type Value struct {
	Type  Type
	Str   string
	Int   int64
	Array []Value
	Null  bool
}

var ErrProtocol = errors.New("erro de protocolo")

const (
	maxBulkLen  = 512 * 1024 * 1024
	maxArrayLen = 1024 * 1024
)

type Reader struct {
	rd *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{rd: bufio.NewReader(r)}
}

func (r *Reader) Read() (Value, error) {
	prefix, err := r.rd.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch Type(prefix) {
	case SimpleString, Error:
		line, err := r.readLine()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: Type(prefix), Str: line}, nil
	case Integer:
		n, err := r.readInt()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: Integer, Int: n}, nil
	case BulkString:
		return r.readBulk()
	case Array:
		return r.readArray()
	default:
		return Value{}, fmt.Errorf("%w: tipo desconhecido %q", ErrProtocol, prefix)
	}
}

func (r *Reader) readLine() (string, error) {
	line, err := r.rd.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", fmt.Errorf("%w: linha sem \\r\\n", ErrProtocol)
	}
	return line[:len(line)-2], nil
}

func (r *Reader) readInt() (int64, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: inteiro inválido %q", ErrProtocol, line)
	}
	return n, nil
}

func (r *Reader) readBulk() (Value, error) {
	n, err := r.readInt()
	if err != nil {
		return Value{}, err
	}
	if n == -1 {
		return Value{Type: BulkString, Null: true}, nil
	}
	if n < 0 || n > maxBulkLen {
		return Value{}, fmt.Errorf("%w: tamanho de bulk inválido %d", ErrProtocol, n)
	}

	buf := make([]byte, n+2)
	if _, err := io.ReadFull(r.rd, buf); err != nil {
		return Value{}, err
	}
	if buf[n] != '\r' || buf[n+1] != '\n' {
		return Value{}, fmt.Errorf("%w: bulk sem \\r\\n no final", ErrProtocol)
	}
	return Value{Type: BulkString, Str: string(buf[:n])}, nil
}

func (r *Reader) readArray() (Value, error) {
	n, err := r.readInt()
	if err != nil {
		return Value{}, err
	}
	if n == -1 {
		return Value{Type: Array, Null: true}, nil
	}
	if n < 0 || n > maxArrayLen {
		return Value{}, fmt.Errorf("%w: tamanho de array inválido %d", ErrProtocol, n)
	}

	items := make([]Value, 0, n)
	for i := int64(0); i < n; i++ {
		item, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		items = append(items, item)
	}
	return Value{Type: Array, Array: items}, nil
}

func (v Value) Args() ([]string, error) {
	if v.Type != Array || v.Null || len(v.Array) == 0 {
		return nil, fmt.Errorf("%w: comando deve ser um array não vazio", ErrProtocol)
	}
	args := make([]string, len(v.Array))
	for i, item := range v.Array {
		if item.Type != BulkString || item.Null {
			return nil, fmt.Errorf("%w: argumento %d não é bulk string", ErrProtocol, i)
		}
		args[i] = item.Str
	}
	return args, nil
}
