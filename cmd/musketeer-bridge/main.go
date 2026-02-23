package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"musketeer-bridge/internal/config"
	"musketeer-bridge/internal/httpapi"
	"musketeer-bridge/internal/logstore"
	"musketeer-bridge/internal/registry"
)

func usage() string {
	return "Usage:\n  musketeer-bridge serve\n  musketeer-bridge help\n  musketeer-bridge --help\n"
}

func serve() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.RunsDir, 0o755); err != nil {
		return err
	}
	reg, err := registry.Load(cfg.RegistryDir)
	if err != nil {
		return err
	}
	api := &httpapi.API{Cfg: cfg, Reg: reg, Log: logstore.LogWriter{RunsDir: cfg.RunsDir}}

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}
	log.Printf("listening on %s", cfg.ListenAddr)

	srv := &http.Server{Handler: api}
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage())
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "--help", "-h", "help":
		fmt.Print(usage())
		os.Exit(0)
	case "serve":
		if len(os.Args) > 2 {
			a := os.Args[2]
			if a == "--help" || a == "-h" || a == "help" {
				fmt.Print("Usage:\n  musketeer-bridge serve\n")
				os.Exit(0)
			}
			fmt.Fprint(os.Stderr, usage())
			os.Exit(2)
		}
		if err := serve(); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprint(os.Stderr, usage())
		os.Exit(2)
	}
}
