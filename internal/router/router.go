package router

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
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
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, rcv interface{}) {
		writeUnexpectedError(w, r, fmt.Errorf("panic: %s", rcv))
	}
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	sh := newSurfingHandler(ss)
	router.GET("/v1/spots", sh.spots)
	router.GET("/v1/spots/:"+paramKeySpotID, sh.spot)

	mh := newManagementHandler(ms)
	router.GET("/management/v1/spots", mh.spots)
	router.GET("/management/v1/spots/:"+paramKeySpotID, mh.spot)
	router.POST("/management/v1/spots", mh.createSpot)
	router.PATCH("/management/v1/spots/:"+paramKeySpotID, mh.updateSpot)
	router.DELETE("/management/v1/spots/:"+paramKeySpotID, mh.deleteSpot)
	router.GET("/management/v1/geo/location", mh.location)

	return withLogger(l, router)
}

func withLogger(l *logrus.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO can be improved by setting fields with request details.
		if l != nil {
			r = r.WithContext(logging.ContextWith(r.Context(), logrus.NewEntry(l)))
		}

		next.ServeHTTP(w, r)
	})
}
