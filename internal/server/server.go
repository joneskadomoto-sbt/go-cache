package server

import (
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/joneskadomoto-sbt/go-cache/internal/resp"
)

type handlerFunc func(args []string, w *resp.Writer)

type Server struct {
	addr     string
	logger   *slog.Logger
	commands map[string]handlerFunc
}

func New(addr string, logger *slog.Logger) *Server {
	s := &Server{addr: addr, logger: logger}
	s.commands = map[string]handlerFunc{
		"PING": s.cmdPing,
		"ECHO": s.cmdEcho,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.logger.Info("servidor ouvindo", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.logger.Error("erro ao aceitar conexão", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	s.logger.Info("cliente conectado", "remote", remote)

	reader := resp.NewReader(conn)
	writer := resp.NewWriter(conn)

	for {
		value, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.Info("cliente desconectado", "remote", remote)
				return
			}
			s.logger.Error("erro lendo comando", "remote", remote, "err", err)
			if errors.Is(err, resp.ErrProtocol) {
				writer.WriteError("ERR " + err.Error())
				writer.Flush()
			}
			return
		}

		args, err := value.Args()
		if err != nil {
			s.logger.Error("comando inválido", "remote", remote, "err", err)
			writer.WriteError("ERR " + err.Error())
			writer.Flush()
			return
		}

		s.logger.Debug("comando recebido", "remote", remote, "args", args)
		s.dispatch(args, writer)

		if err := writer.Flush(); err != nil {
			s.logger.Error("erro ao enviar resposta", "remote", remote, "err", err)
			return
		}
	}
}
