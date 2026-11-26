package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/joneskadomoto-sbt/go-cache/internal/server"
)

func main() {
	addr := flag.String("addr", ":6380", "endereço TCP em que o servidor escuta")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	srv := server.New(*addr, logger)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Servidor encerrou com erro", "err", err)
		os.Exit(1)
	}
}
