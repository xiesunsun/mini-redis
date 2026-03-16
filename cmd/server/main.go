package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xiesunsun/mini-redis/internal/command"
	"github.com/xiesunsun/mini-redis/internal/expiry"
	"github.com/xiesunsun/mini-redis/internal/network"
	"github.com/xiesunsun/mini-redis/internal/persistence"
	"github.com/xiesunsun/mini-redis/internal/store"
	"github.com/xiesunsun/mini-redis/internal/types"
)

const (
	defaultListenAddr      = ":6379"
	defaultAOFPath         = "appendonly.aof"
	defaultCleanupInterval = 100 * time.Millisecond
)

func main() {
	srv, cleanup, err := buildServer(defaultListenAddr, defaultAOFPath, defaultCleanupInterval)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		if err != nil {
			_ = cleanup()
			log.Fatalf("server stopped with error: %v", err)
		}
		if err := cleanup(); err != nil {
			log.Fatalf("shutdown failed: %v", err)
		}
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig)
		if err := cleanup(); err != nil {
			log.Fatalf("shutdown failed: %v", err)
		}
		if err := <-errCh; err != nil {
			log.Fatalf("server close returned error: %v", err)
		}
	}
}

func buildServer(addr, aofPath string, cleanupInterval time.Duration) (*network.Server, func() error, error) {
	s := store.New()
	aof, err := persistence.New(aofPath)
	if err != nil {
		return nil, nil, err
	}

	if err := replayAOF(aof, s); err != nil {
		_ = aof.Close()
		return nil, nil, err
	}

	ctx := &command.Context{
		Store: s,
		AOF:   aof,
	}
	srv := network.NewServer(addr, ctx)
	stopCleaner := expiry.StartCleaner(s, cleanupInterval)

	var once sync.Once
	var cleanupErr error
	cleanup := func() error {
		once.Do(func() {
			stopCleaner()
			cleanupErr = errors.Join(srv.Close(), aof.Close())
		})
		return cleanupErr
	}

	return srv, cleanup, nil
}

func replayAOF(aof *persistence.AOF, s *store.Store) error {
	cmds, err := aof.Replay()
	if err != nil {
		return err
	}

	replayCtx := &command.Context{
		Store: s,
		AOF:   nil,
	}
	router := command.NewRouter(replayCtx)

	for i, cmd := range cmds {
		resp := router.Dispatch(types.Command{Name: cmd.Name, Args: cmd.Args})
		if strings.HasPrefix(resp, "-") {
			return fmt.Errorf("replay command #%d %s failed: %s", i, cmd.Name, strings.TrimSpace(resp))
		}
	}
	return nil
}
