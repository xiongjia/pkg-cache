package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/xiongjia/pkg-cache/pkg/config"
	"github.com/xiongjia/pkg-cache/pkg/proxy"
)

var log *slog.Logger

func init() {
	log = slog.Default()
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	log.Info("starting pkg_cache",
		"port", cfg.Server.Port,
		"cache_dir", cfg.Cache.Directory,
		"cache_size_mb", cfg.Cache.MaxSizeMB)
	if cfg.ParentProxy != "" {
		log.Info("using parent proxy", "proxy", cfg.ParentProxy)
	}

	p := proxy.New(cfg)
	handler := http.HandlerFunc(p.ServeHTTP)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), handler); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
