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
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/update"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/version"
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
	updater := update.NewChecker(version.Display(), dataDir)
	srv := server.New(st, skinStore, h, obs, info, webassets.FS, updater)

	httpServer := &http.Server{Handler: srv.Handler()}
	go func() {
		log.Printf("listening on %s (version %s)", info.BaseURL, version.Display())
		log.Printf("overlay: %s", info.OverlayURL)
		log.Printf("admin:   %s", info.AdminURL)
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	quitOnce := func() {
		stop()
		shutdown(httpServer)
		systray.Quit()
	}
	srv.OnQuit = func() {
		quitOnce()
		os.Exit(0)
	}

	hotkeydrop.Start(ctx, dataDir, func(a hotkeydrop.Action) {
		if a == hotkeydrop.Quit {
			stop()
			return
		}
		srv.ApplyHotkey(string(a))
	})

	updateCh := make(chan tray.UpdateNotice, 2)
	go func() {
		<-ctx.Done()
		shutdown(httpServer)
		systray.Quit()
	}()

	go func() {
		time.Sleep(2 * time.Second)
		info := updater.Check(false)
		if info.Available {
			log.Printf("update available: %s → %s", info.Current, info.Latest)
			select {
			case updateCh <- tray.UpdateNotice{Available: true, Latest: info.Latest}:
			default:
			}
		}
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info := updater.Check(true)
				if info.Available {
					select {
					case updateCh <- tray.UpdateNotice{Available: true, Latest: info.Latest}:
					default:
					}
				}
			}
		}
	}()

	tray.Run(tray.Hooks{
		AdminURL: info.AdminURL,
		Updates:  updateCh,
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
