package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/management"
	"github.com/ztimes2/tolqin/internal/surfing"
)

const (
	paramKeySpotID = "spot_id"
)

func New(ss *surfing.Service, ms *management.Service, l *logrus.Logger) http.Handler {
	return newRouter(ss, ms, l)
}

func newRouter(ss surfingService, ms managementService, l *logrus.Logger) http.Handler {
	router := chi.NewRouter()

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	router.Use(
		withLogger(l),
		withPanicRecoverer,
	)

	sh := newSurfingHandler(ss)
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
				r = r.WithContext(logging.ContextWith(r.Context(), logrus.NewEntry(l)))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func withPanicRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rcv := recover(); rcv != nil && rcv != http.ErrAbortHandler {
			// TODO add stack trace
			writeUnexpectedError(w, r, fmt.Errorf("panic: %s", rcv))
			return
		}

		next.ServeHTTP(w, r)
	})
}
