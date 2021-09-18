package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/service/management"
)

type managementService interface {
	Spot(id string) (management.Spot, error)
	Spots(management.SpotsParams) ([]management.Spot, error)
	CreateSpot(management.CreateSpotParams) (management.Spot, error)
	UpdateSpot(management.UpdateSpotParams) (management.Spot, error)
	DeleteSpot(id string) error
	Location(geo.Coordinates) (geo.Location, error)
}

func fromManagementSpot(s management.Spot) spotResponse {
	return spotResponse{
		ID:          s.ID,
		Name:        s.Name,
		Latitude:    s.Location.Coordinates.Latitude,
		Longitude:   s.Location.Coordinates.Longitude,
		Locality:    s.Location.Locality,
		CountryCode: s.Location.CountryCode,
	}
}

type managementHandler struct {
	service managementService
}

func newManagementHandler(s managementService) *managementHandler {
	return &managementHandler{
		service: s,
	}
}

func (h *managementHandler) spot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, paramKeySpotID)

	spot, err := h.service.Spot(id)
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, paramKeySpotID, "Must be a non empty string.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, management.ErrNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, fromManagementSpot(spot))
}

func (h *managementHandler) spots(w http.ResponseWriter, r *http.Request) {
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

	spots, err := h.service.Spots(management.SpotsParams{
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
				f.Is(e, management.ErrInvalidSearchQuery, "query", "Must not exceed character limit.")
				f.Is(e, management.ErrInvalidCountryCode, "country", "Must be a valid ISO-2 country code.")
				f.Is(e, management.ErrInvalidNorthEastLatitude, "ne_lat", "Must be a valid latitude.")
				f.Is(e, management.ErrInvalidNorthEastLongitude, "ne_lon", "Must be a valid longitude.")
				f.Is(e, management.ErrInvalidSouthWestLatitude, "sw_lat", "Must be a valid latitude.")
				f.Is(e, management.ErrInvalidSouthWestLongitude, "sw_lon", "Must be a valid longitude.")
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
		resp.Items[i] = fromManagementSpot(s)
	}

	httputil.WriteOK(w, r, resp)
}

func (h *managementHandler) createSpot(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string  `json:"name"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Locality    string  `json:"locality"`
		CountryCode string  `json:"country_code"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WritePayloadError(w, r)
		return
	}

	spot, err := h.service.CreateSpot(management.CreateSpotParams{
		Name: payload.Name,
		Location: geo.Location{
			Coordinates: geo.Coordinates{
				Latitude:  payload.Latitude,
				Longitude: payload.Longitude,
			},
			Locality:    payload.Locality,
			CountryCode: payload.CountryCode,
		},
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotName, "name", "Must be a non empty string.")
				f.Is(e, management.ErrInvalidCountryCode, "country_code", "Must be a valid ISO-2 country code.")
				f.Is(e, management.ErrInvalidLocality, "locality", "Must be a non empty string.")
				f.Is(e, management.ErrInvalidLatitude, "latitude", "Must be a valid latitude.")
				f.Is(e, management.ErrInvalidLongitude, "longitude", "Must be a valid longitude.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteCreated(w, r, fromManagementSpot(spot))
}

func (h *managementHandler) updateSpot(w http.ResponseWriter, r *http.Request) {
	spotID := chi.URLParam(r, paramKeySpotID)

	var payload struct {
		Name        *string  `json:"name"`
		Latitude    *float64 `json:"latitude"`
		Longitude   *float64 `json:"longitude"`
		Locality    *string  `json:"locality"`
		CountryCode *string  `json:"country_code"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httputil.WritePayloadError(w, r)
		return
	}

	spot, err := h.service.UpdateSpot(management.UpdateSpotParams{
		ID:          spotID,
		Name:        payload.Name,
		Latitude:    payload.Latitude,
		Longitude:   payload.Longitude,
		Locality:    payload.Locality,
		CountryCode: payload.CountryCode,
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, paramKeySpotID, "Must be a non empty string.")
				f.Is(e, management.ErrInvalidSpotName, "name", "Must be a non empty string.")
				f.Is(e, management.ErrInvalidCountryCode, "country_code", "Must be a valid ISO-2 country code.")
				f.Is(e, management.ErrInvalidLocality, "locality", "Must be a non empty string.")
				f.Is(e, management.ErrInvalidLatitude, "latitude", "Must be a valid latitude.")
				f.Is(e, management.ErrInvalidLongitude, "longitude", "Must be a valid longitude.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, management.ErrNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		if errors.Is(err, management.ErrNothingToUpdate) {
			httputil.WriteValidationError(w, r, "Nothing to update.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, fromManagementSpot(spot))
}

func (h *managementHandler) deleteSpot(w http.ResponseWriter, r *http.Request) {
	spotID := chi.URLParam(r, paramKeySpotID)

	if err := h.service.DeleteSpot(spotID); err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, paramKeySpotID, "Must be a non empty string.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, management.ErrNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteNoContent(w, r)
}

func (h *managementHandler) location(w http.ResponseWriter, r *http.Request) {
	latitude, err := httputil.QueryParamFloat(r, "lat")
	if err != nil {
		httputil.WriteFieldError(w, r, "lat", "Must be a valid latitude.")
		return
	}

	longitude, err := httputil.QueryParamFloat(r, "lon")
	if err != nil {
		httputil.WriteFieldError(w, r, "lon", "Must be a valid longitude.")
		return
	}

	l, err := h.service.Location(geo.Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidLatitude, "lat", "Must be a valid latitude.")
				f.Is(e, management.ErrInvalidLongitude, "lon", "Must be a valid longitude.")
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, management.ErrNotFound) {
			httputil.WriteNotFoundError(w, r, "Location was not found.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	resp := locationResponse{
		Latitude:    l.Coordinates.Latitude,
		Longitude:   l.Coordinates.Longitude,
		Locality:    l.Locality,
		CountryCode: l.CountryCode,
	}

	httputil.WriteOK(w, r, resp)
}
