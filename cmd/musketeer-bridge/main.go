package main

import (
	"log"
	"net/http"
	"os"

	"musketeer-bridge/internal/config"
	"musketeer-bridge/internal/httpapi"
	"musketeer-bridge/internal/logstore"
	"musketeer-bridge/internal/registry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.MkdirAll(cfg.RunsDir, 0o755)
	reg, err := registry.Load(cfg.RegistryDir)
	if err != nil {
		log.Fatal(err)
	}
	api := &httpapi.API{Cfg: cfg, Reg: reg, Log: logstore.LogWriter{RunsDir: cfg.RunsDir}}
	log.Printf("listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, api); err != nil {
		log.Fatal(err)
	}
}
