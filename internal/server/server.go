package server

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
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

	s.logger.Info("Servidor ouvindo", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.logger.Error("Erro ao aceitar conexão", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	s.logger.Info("cliente conectado", "remote", remote)

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			s.logger.Info("Cliente desconectado", "remote", remote, "motivo", err)
			return
		}
		if _, err := conn.Write([]byte(line)); err != nil {
			s.logger.Error("erro ao escrever", "remote", remote, "err", err)
			return
		}
	}
}
