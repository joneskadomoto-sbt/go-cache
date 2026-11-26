package server

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/joneskadomoto-sbt/go-cache/internal/resp"
)

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(":0", logger)
}

func TestDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"ping sem argumento", []string{"PING"}, "+PONG\r\n"},
		{"ping com argumento", []string{"PING", "oi"}, "$2\r\noi\r\n"},
		{"ping minúsculo", []string{"ping"}, "+PONG\r\n"},
		{"ping com argumentos demais", []string{"PING", "a", "b"}, "-ERR wrong number of arguments for 'ping' command\r\n"},
		{"echo", []string{"ECHO", "olá mundo"}, "$10\r\nolá mundo\r\n"},
		{"echo sem argumento", []string{"ECHO"}, "-ERR wrong number of arguments for 'echo' command\r\n"},
		{"comando desconhecido", []string{"FOO"}, "-ERR unknown command 'FOO'\r\n"},
	}

	s := newTestServer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := resp.NewWriter(&buf)

			s.dispatch(tt.args, w)
			w.Flush()

			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
