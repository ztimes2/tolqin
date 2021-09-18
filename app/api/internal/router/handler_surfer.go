package router

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/service/surfer"
)

type surferService interface {
	Spot(id string) (surfer.Spot, error)
	Spots(surfer.SpotsParams) ([]surfer.Spot, error)
}

func fromSurferSpot(s surfer.Spot) spotResponse {
	return spotResponse{
		ID:          s.ID,
		Name:        s.Name,
		Latitude:    s.Location.Coordinates.Latitude,
		Longitude:   s.Location.Coordinates.Longitude,
		Locality:    s.Location.Locality,
		CountryCode: s.Location.CountryCode,
	}
}

type surferHandler struct {
	service surferService
}

func newSurferHandler(s surferService) *surferHandler {
	return &surferHandler{
		service: s,
	}
}

func (h *surferHandler) spot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, paramKeySpotID)

	spot, err := h.service.Spot(id)
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, surfer.ErrInvalidSpotID, paramKeySpotID, "Must be a non empty string.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, surfer.ErrNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, fromSurferSpot(spot))
}

func (h *surferHandler) spots(w http.ResponseWriter, r *http.Request) {
	limit, err := httputil.QueryParamInt(r, "limit")
	if err != nil && !errors.Is(err, httputil.ErrEmptyParam) {
		httputil.WriteFieldError(w, r, "limit", "Must be a valid integer.")
		return
	}

	offset, err := httputil.QueryParamInt(r, "offset")
	if err != nil && !errors.Is(err, httputil.ErrEmptyParam) {
		httputil.WriteFieldError(w, r, "offset", "Must be a valid integer.")
		return
	}

	countryCode := httputil.QueryParam(r, "country")

	query := httputil.QueryParam(r, "query")

	bounds, vErr := parseBounds(
		httputil.QueryParam(r, "ne_lat"),
		httputil.QueryParam(r, "ne_lon"),
		httputil.QueryParam(r, "sw_lat"),
		httputil.QueryParam(r, "sw_lon"),
	)
	if vErr != nil {
		f := httputil.NewFields()
		for _, e := range vErr.Errors() {
			f.Is(e, errInvalidNorthEastLatitude, "ne_lat", "Must be a valid latitude.")
			f.Is(e, errInvalidNorthEastLongitude, "ne_lon", "Must be a valid longitude.")
			f.Is(e, errInvalidSouthWestLatitude, "sw_lat", "Must be a valid latitude.")
			f.Is(e, errInvalidSouthWestLongitude, "sw_lon", "Must be a valid longitude.")
		}
		httputil.WriteFieldErrors(w, r, f)
		return
	}

	spots, err := h.service.Spots(surfer.SpotsParams{
		Limit:       limit,
		Offset:      offset,
		CountryCode: countryCode,
		Query:       query,
		Bounds:      bounds,
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, surfer.ErrInvalidSearchQuery, "query", "Must not exceed character limit.")
				f.Is(e, surfer.ErrInvalidCountryCode, "country", "Must be a valid ISO-2 country code.")
				f.Is(e, surfer.ErrInvalidNorthEastLatitude, "ne_lat", "Must be a valid latitude.")
				f.Is(e, surfer.ErrInvalidNorthEastLongitude, "ne_lon", "Must be a valid longitude.")
				f.Is(e, surfer.ErrInvalidSouthWestLatitude, "sw_lat", "Must be a valid latitude.")
				f.Is(e, surfer.ErrInvalidSouthWestLongitude, "sw_lon", "Must be a valid longitude.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	resp := spotsResponse{
		Items: make([]spotResponse, len(spots)),
	}

	for i, s := range spots {
		resp.Items[i] = fromSurferSpot(s)
	}

	httputil.WriteOK(w, r, resp)
}
