package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/twitsprout/tools"
	httputils "github.com/twitsprout/tools/http"
	"github.com/twitsprout/tools/runtime"
)

// LifeCycle manages the running of one or more processes, returning when one
// of the processes exits. When a process exits, the LifeCycle's context will
// be cancelled. The Wait method will block until all processes exit or a
// timeout is reached.
type LifeCycle struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger tools.Logger

	wg   sync.WaitGroup
	once sync.Once
	mu   sync.Mutex
	err  error
}

// New returns a new LifeCycle using the provided parent context and logger. A
// new context is also returned, which will be cancelled when any of the
// LifeCycle's processes exit.
func New(ctx context.Context, logger tools.Logger) (*LifeCycle, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &LifeCycle{
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}, ctx
}

// Start executes the provided function in a new goroutine, returning a non-nil
// error explaining why it returned. When the context provided when the
// LifeCycle was created is cancelled, the process should return ASAP.
func (lc *LifeCycle) Start(name string, fn func() error) {
	lc.wg.Add(1)
	go func() {
		lc.logger.Info("starting process",
			"name", name,
		)
		var err error
		defer func() {
			lc.wg.Done()
			if r := recover(); r != nil {
				lc.logger.Error("recovered from lifecycle panic",
					"name", name,
					"details", r,
					"stacktrace", runtime.Stacktrace(0),
				)
				err = fmt.Errorf("recovered from panic: %v", r)
			}
			if err == nil {
				err = errors.New("nil error returned from process")
			}
			lc.logger.Info("process shutdown",
				"name", name,
				"details", err.Error(),
			)
			err = fmt.Errorf("%s: %s", name, err.Error())
			lc.setError(err)
		}()
		err = fn()
	}()
}

// StartServer starts the provided server, gracefully shutting it down when the
// LifeCycle's context is cancelled.
func (lc *LifeCycle) StartServer(s *http.Server) {
	name := fmt.Sprintf("http server '%s'", s.Addr)
	lc.Start(name, func() error {
		go func() {
			// Sleep for a second while the server actually starts.
			time.Sleep(time.Second)
			<-lc.ctx.Done()
			_ = s.Shutdown(context.Background())
		}()
		return httputils.ListenAndServe(s, 30*time.Second)
	})
}

// StartSignals listens to the provided OS signals and will exit when a signal
// is received.
func (lc *LifeCycle) StartSignals(signals ...os.Signal) {
	lc.Start("signal listener", func() error {
		chSig := make(chan os.Signal, 1)
		signal.Notify(chSig, signals...)
		select {
		case <-lc.ctx.Done():
			return lc.ctx.Err()
		case sig := <-chSig:
			return fmt.Errorf("received os signal: %s", sig.String())
		}
	})
}

// Wait will block until all started processes exit, or the LifeCycle's context
// is cancelled and the graceful timeout is reached. It will return the error
// indicating why the LifeCycle was closed.
func (lc *LifeCycle) Wait(gracefulTimeout time.Duration) error {
	// Wait for the context to be cancelled.
	<-lc.ctx.Done()
	lc.setError(lc.ctx.Err())

	// Send an error on this channel when the services exit.
	chIsShutdown := make(chan struct{})
	go func() {
		lc.wg.Wait()
		close(chIsShutdown)
	}()

	// Wait for services to gracefully shutdown.
	var err error
	select {
	case <-chIsShutdown:
		err = lc.getError()
		lc.logger.Info("all processes shutdown",
			"reason", err.Error(),
		)
	case <-time.After(gracefulTimeout):
		err = lc.getError()
		lc.logger.Warn("timeout reached while waiting for processes to gracefully shutdown",
			"timeout", gracefulTimeout,
			"reason", err.Error(),
		)
	}
	lc.logger.Info("goodbye")
	return err
}

func (lc *LifeCycle) getError() error {
	lc.mu.Lock()
	err := lc.err
	lc.mu.Unlock()
	return err
}

func (lc *LifeCycle) setError(err error) {
	lc.once.Do(func() {
		lc.mu.Lock()
		lc.err = err
		lc.mu.Unlock()
		lc.cancel()
	})
}
