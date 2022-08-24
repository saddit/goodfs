package graceful

import (
	"common/logs"
	"common/util"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ServableServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

var log = logs.New("server")

func ListenAndServe(srvs ...ServableServer) {
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	for _, s := range srvs {
		go func(srv ServableServer) {
			if err := srv.ListenAndServe(); err != nil {
				log.Errorf("listen server error: %s", err)
			}
		}(s)
	}
	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Infoln("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dg := util.NewDoneGroup()
	for _, s := range srvs {
		dg.Todo()
		go func (srv ServableServer) {
			defer dg.Done()
			if err := srv.Shutdown(ctx); err != nil {
				log.Errorf("Server shutdown err: %s", err)
			}
		}(s)
	}
	dg.Wait()
	log.Infoln("Server shudown over")
}
