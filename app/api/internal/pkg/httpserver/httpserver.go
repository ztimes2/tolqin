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

type Server struct {
	logger          *logrus.Logger
	shutdownTimeout time.Duration

	server   *http.Server
	isClosed *syncBool
	closeCh  chan struct{}
}

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

type Option func(*Server)

func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

func WithLogger(l *logrus.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

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

func (s *Server) Close() error {
	if s.isClosed.get() {
		return http.ErrServerClosed
	}
	s.closeCh <- struct{}{}
	return nil
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
