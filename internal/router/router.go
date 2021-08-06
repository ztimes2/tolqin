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

func New(s *surfing.Service, l *logrus.Logger) http.Handler {
	return newRouter(s, l)
}

type service interface {
	Spot(id string) (surfing.Spot, error)
	Spots(surfing.SpotsParams) ([]surfing.Spot, error)
	CreateSpot(surfing.CreateSpotParams) (surfing.Spot, error)
	UpdateSpot(surfing.UpdateSpotParams) (surfing.Spot, error)
	DeleteSpot(id string) error
}

func newRouter(s service, l *logrus.Logger) http.Handler {
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, rcv interface{}) {
		writeUnexpectedError(w, r, fmt.Errorf("panic: %s", rcv))
	}
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	h := newHandler(s)

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
		if l != nil {
			r = r.WithContext(logging.ContextWith(r.Context(), logrus.NewEntry(l)))
		}

		next.ServeHTTP(w, r)
	})
}
