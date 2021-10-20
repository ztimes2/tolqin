package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	serviceauth "github.com/ztimes2/tolqin/app/api/internal/api/service/auth"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/jwt"
	"github.com/ztimes2/tolqin/app/api/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/pkg/log"
)

const (
	paramKeySpotID = "spot_id"
)

// New returns an HTTP router that serves various APIs of the application.
func New(
	as *serviceauth.Service,
	ss *surfer.Service,
	ms *management.Service,
	j *jwt.EncodeDecoder,
	l *logrus.Logger) http.Handler {

	return newRouter(as, ss, ms, j, l)
}

func newRouter(
	as authService,
	ss surferService,
	ms managementService,
	j *jwt.EncodeDecoder,
	l *logrus.Logger) http.Handler {

	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	router.Use(
		withLogger(l),
		withPanicRecoverer,
		withJWTClaims(j),
	)

	router.Get("/health", handleHealthCheck)

	ah := newAuthHandler(as)
	router.Post("/auth/v1/token", ah.token)

	sh := newSurferHandler(ss)
	router.Get("/v1/spots", sh.spots)
	router.Get("/v1/spots/{"+paramKeySpotID+"}", sh.spot)

	mh := newManagementHandler(ms)
	router.Get("/management/v1/spots", mh.spots)
	router.Get("/management/v1/spots/{"+paramKeySpotID+"}", mh.spot)
	router.Post("/management/v1/spots", mh.createSpot)
	router.Patch("/management/v1/spots/{"+paramKeySpotID+"}", mh.updateSpot)
	router.Delete("/management/v1/spots/{"+paramKeySpotID+"}", mh.deleteSpot)
	router.Get("/management/v1/geo/location", mh.location)

	return router
}

func withLogger(l *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO can be improved by setting fields with request details.
			if l != nil {
				r = r.WithContext(log.ContextWith(r.Context(), logrus.NewEntry(l)))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func withPanicRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rcv := recover(); rcv != nil && rcv != http.ErrAbortHandler {
			// TODO add stack trace
			httputil.WriteUnexpectedError(w, r, fmt.Errorf("panic: %s", rcv))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withJWTClaims(j *jwt.EncodeDecoder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := httputil.BearerAuthHeader(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := j.DecodeJWT(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(jwt.ContextWith(r.Context(), claims))

			next.ServeHTTP(w, r)
		})
	}
}
