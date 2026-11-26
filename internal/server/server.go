package server

import (
	"errors"
	"io"
	"log/slog"
	"net"

	"github.com/joneskadomoto-sbt/go-cache/internal/resp"
)

type Server struct {
	addr   string
	logger *slog.Logger
}

func New(addr string, logger *slog.Logger) *Server {
	return &Server{addr: addr, logger: logger}
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
	for {
		value, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.Info("cliente desconectado", "remote", remote)
			} else {
				s.logger.Error("erro lendo comando", "remote", remote, "err", err)
			}
			return
		}

		args, err := value.Args()
		if err != nil {
			s.logger.Error("comando inválido", "remote", remote, "err", err)
			return
		}

		s.logger.Info("comando recebido", "remote", remote, "args", args)
	}
}
