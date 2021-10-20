package httpserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Server is a lightweight wrapper around a standard http.Server that additionally
// enables automatic graceful shutdown during server errors and logging.
type Server struct {
	logger          *logrus.Logger
	shutdownTimeout time.Duration

	server   *http.Server
	isClosed *syncBool
	closeCh  chan struct{}
}

// New returns a new *Server using the given port, HTTP handler, and other options.
//
// By default, the server is shipped without a shutdown timeout and a default
// *logrus.Logger unless they are overwritten via opts.
func New(port string, h http.Handler, opts ...Option) *Server {
	s := &Server{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: h,
			// TODO configure timeouts
		},
		logger:   logrus.StandardLogger(),
		closeCh:  make(chan struct{}, 1),
		isClosed: newBool(false),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Option is an optional function for Server.
type Option func(*Server)

// WithShutdownTimeout sets a shutdown timeout for Server as long as the duration
// is greater than 0. Values less than 1 are interpreted as if no shutdown is desired.
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

// WithLogger sets a custom *logrus.Logger for Server.
func WithLogger(l *logrus.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

// ListenAndServe spins up a server and starts accepting/serving HTTP requests.
//
// It keeps running until a server error is caught, syscall.SIGTERM/syscall.SIGINT
// signals are triggered, or Close() method is envoked. The server gracefully
// shuts itself down if one of the listed cases happen.
func (s *Server) ListenAndServe() error {
	if s.isClosed.get() {
		return http.ErrServerClosed
	}

	errCh := make(chan error)
	go func() {
		s.logger.Infof("server is listening on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)

	var err error
	select {
	case err = <-errCh:
	case <-stopCh:
		s.logger.Info("received termination signal")
	case <-s.closeCh:
	}

	s.logger.Info("shutting down server")

	defer s.server.Close()
	if sdErr := s.shutdown(); sdErr != nil {
		s.logger.WithError(sdErr).Errorf("failed to gracefully shut down server: %v", sdErr)
	}

	s.isClosed.set(true)

	return err
}

func (s *Server) shutdown() error {
	ctx := context.Background()
	cancel := func() {}

	if s.shutdownTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.shutdownTimeout)
	}
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

// Close tells the server to gracefully shut itself down if it's running.
//
// It's not mandatory to envoke this method unless a manual shutdown is desired
// which is probably rare since the server is capable of automatically shutting
// itself down during the majority of the practical cases.
func (s *Server) Close() {
	if !s.isClosed.get() {
		s.closeCh <- struct{}{}
	}
}

type syncBool struct {
	b  bool
	mu sync.Mutex
}

func newBool(b bool) *syncBool {
	return &syncBool{
		b: b,
	}
}

func (sb *syncBool) get() bool {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.b
}

func (sb *syncBool) set(b bool) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.b = b
}
