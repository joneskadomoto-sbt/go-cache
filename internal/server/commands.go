package server

import (
	"fmt"
	"strings"

	"github.com/joneskadomoto-sbt/go-cache/internal/resp"
)

func (s *Server) dispatch(args []string, w *resp.Writer) {
	name := strings.ToUpper(args[0])

	handler, ok := s.commands[name]

	if !ok {
		w.WriteError(fmt.Sprintf("ERR unknown command '%s'", args[0]))
		return
	}
	handler(args[1:], w)
}

func wrongArgs(w *resp.Writer, cmd string) {
	w.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", cmd))
}

func (s *Server) cmdPing(args []string, w *resp.Writer) {
	switch len(args) {
	case 0:
		w.WriteSimpleString("PONG")

	case 1:
		w.WriteBulkString(args[0])

	default:
		wrongArgs(w, "ping")
	}
}

func (s *Server) cmdEcho(args []string, w *resp.Writer) {
	if len(args) != 1 {
		wrongArgs(w, "echo")
		return
	}
	w.WriteBulkString(args[0])
}
