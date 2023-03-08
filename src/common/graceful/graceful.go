package graceful

import (
	"common/logs"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

var log = logs.New("graceful")

func ListenAndServe(ctx context.Context, srvs ...Server) {
	if ctx == nil {
		ctx = context.Background()
	}
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	for _, s := range srvs {
		go func(srv Server) {
			defer Recover()
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("listen server error: %s", err)
			}
		}(s)
	}

	defer func() {
		wg := sync.WaitGroup{}
		// The context is used to inform the servers it has 15 seconds to finish the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		for _, s := range srvs {
			wg.Add(1)
			go func(srv Server) {
				defer Recover()
				defer wg.Done()
				if err := srv.Shutdown(ctx); err != nil {
					log.Errorf("server shutdown err: %s", err)
				}
			}(s)
		}
		wg.Wait()
		log.Infoln("server shutdown over")
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-quit:
	case <-ctx.Done():
	}
	log.Infoln("shutting down server...")
}
