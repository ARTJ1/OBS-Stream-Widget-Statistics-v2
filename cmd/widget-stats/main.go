package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hotkeydrop"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hub"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/obsbridge"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/runtimeinfo"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/server"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/skins"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/tray"
	webassets "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/web"
	"github.com/getlantern/systray"
)

func main() {
	dataDir := dataDirPath()
	lock, err := runtimeinfo.AcquireLock(dataDir)
	if err != nil {
		log.Fatalf("single-instance: %v", err)
	}
	defer lock.Close()

	st, err := store.New(dataDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	port, ln, err := runtimeinfo.FindPort(runtimeinfo.DefaultPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	info := runtimeinfo.Build("127.0.0.1", port)
	if err := runtimeinfo.Save(dataDir, info); err != nil {
		log.Fatalf("runtime: %v", err)
	}

	h := hub.New()
	obs := obsbridge.New()
	skinStore, err := skins.New(dataDir)
	if err != nil {
		log.Fatalf("skins: %v", err)
	}
	srv := server.New(st, skinStore, h, obs, info, webassets.FS)

	httpServer := &http.Server{Handler: srv.Handler()}
	go func() {
		log.Printf("listening on %s", info.BaseURL)
		log.Printf("overlay: %s", info.OverlayURL)
		log.Printf("admin:   %s", info.AdminURL)
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OBS Lua drops hk_*.wreq here — no curl/cmd, so no console flash in games.
	hotkeydrop.Start(ctx, dataDir, func(a hotkeydrop.Action) {
		if a == hotkeydrop.Quit {
			stop()
			return
		}
		srv.ApplyHotkey(string(a))
	})

	go func() {
		<-ctx.Done()
		shutdown(httpServer)
		systray.Quit()
	}()

	tray.Run(tray.Hooks{
		AdminURL: info.AdminURL,
		OnQuit: func() {
			stop()
			shutdown(httpServer)
			os.Exit(0)
		},
	})
}

func shutdown(httpServer *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}

func dataDirPath() string {
	if v := os.Getenv("WIDGET_STATS_DATA"); v != "" {
		return v
	}
	exe, err := os.Executable()
	if err != nil {
		return "data"
	}
	return filepath.Join(filepath.Dir(exe), "data")
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix(fmt.Sprintf("[widget-stats] "))
}
