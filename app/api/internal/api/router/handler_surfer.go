package router

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
	"github.com/ztimes2/tolqin/app/api/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

type surferService interface {
	Spot(id string) (surf.Spot, error)
	Spots(surfer.SpotsParams) ([]surf.Spot, error)
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
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, surfer.ErrInvalidSpotID, httputil.NewInvalidField(paramKeySpotID, "Must be a non empty string."))
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, surf.ErrSpotNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, toSpotResponse(spot))
}

func (h *surferHandler) spots(w http.ResponseWriter, r *http.Request) {
	limit, err := httputil.QueryParamInt(r, "limit")
	if err != nil && !errors.Is(err, httputil.ErrParamNotFound) {
		httputil.WriteFieldError(w, r, httputil.NewInvalidField("limit", "Must be a valid integer."))
		return
	}

	offset, err := httputil.QueryParamInt(r, "offset")
	if err != nil && !errors.Is(err, httputil.ErrParamNotFound) {
		httputil.WriteFieldError(w, r, httputil.NewInvalidField("offset", "Must be a valid integer."))
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
		f := httputil.NewInvalidFields()
		for _, e := range vErr.Errors() {
			f.Is(e, errInvalidNorthEastLatitude, httputil.NewInvalidField("ne_lat", "Must be a valid latitude."))
			f.Is(e, errInvalidNorthEastLongitude, httputil.NewInvalidField("ne_lon", "Must be a valid longitude."))
			f.Is(e, errInvalidSouthWestLatitude, httputil.NewInvalidField("sw_lat", "Must be a valid latitude."))
			f.Is(e, errInvalidSouthWestLongitude, httputil.NewInvalidField("sw_lon", "Must be a valid longitude."))
		}
		httputil.WriteFieldErrors(w, r, f)
		return
	}

	spots, err := h.service.Spots(surfer.SpotsParams{
		Limit:       limit,
		Offset:      offset,
		CountryCode: countryCode,
		SearchQuery: query,
		Bounds:      bounds,
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, surfer.ErrInvalidSearchQuery, httputil.NewInvalidField("query", "Must not exceed character limit."))
				f.Is(e, surfer.ErrInvalidCountryCode, httputil.NewInvalidField("country", "Must be a valid ISO-2 country code."))
				f.Is(e, surfer.ErrInvalidNorthEastLatitude, httputil.NewInvalidField("ne_lat", "Must be a valid latitude."))
				f.Is(e, surfer.ErrInvalidNorthEastLongitude, httputil.NewInvalidField("ne_lon", "Must be a valid longitude."))
				f.Is(e, surfer.ErrInvalidSouthWestLatitude, httputil.NewInvalidField("sw_lat", "Must be a valid latitude."))
				f.Is(e, surfer.ErrInvalidSouthWestLongitude, httputil.NewInvalidField("sw_lon", "Must be a valid longitude."))
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
		resp.Items[i] = toSpotResponse(s)
	}

	httputil.WriteOK(w, r, resp)
}
