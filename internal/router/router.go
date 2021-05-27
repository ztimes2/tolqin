package router

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/internal/logging"
	"github.com/ztimes2/tolqin/internal/surfing"
)

const (
	paramKeySpotID = "spot_id"
)

func New(service *surfing.Service, l *logrus.Logger) http.Handler {
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, rcv interface{}) {
		writeUnexpectedError(w, r, fmt.Errorf("panic: %s", rcv))
	}
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	h := newHandler(service)

	router.GET("/spots", h.spots)
	router.GET("/spots/:"+paramKeySpotID, h.spot)
	router.POST("/spots", h.createSpot)
	router.PATCH("/spots/:"+paramKeySpotID, h.updateSpot)
	router.DELETE("/spots/:"+paramKeySpotID, h.deleteSpot)

	return withLogger(l, router)
}

func withLogger(l *logrus.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO can be improved by setting fields with request details.
		entry := logrus.NewEntry(l)

		r = r.WithContext(logging.ContextWith(r.Context(), entry))

		next.ServeHTTP(w, r)
	})
}
