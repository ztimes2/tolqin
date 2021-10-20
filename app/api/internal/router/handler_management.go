package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/httputil"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/surf"
)

type managementService interface {
	Spot(ctx context.Context, id string) (surf.Spot, error)
	Spots(context.Context, management.SpotsParams) ([]surf.Spot, error)
	CreateSpot(context.Context, management.CreateSpotParams) (surf.Spot, error)
	UpdateSpot(context.Context, management.UpdateSpotParams) (surf.Spot, error)
	DeleteSpot(ctx context.Context, id string) error
	Location(context.Context, geo.Coordinates) (geo.Location, error)
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

	spot, err := h.service.Spot(r.Context(), id)
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, httputil.NewInvalidField(paramKeySpotID, "Must be a non empty string."))
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

func (h *managementHandler) spots(w http.ResponseWriter, r *http.Request) {
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

	spots, err := h.service.Spots(r.Context(), management.SpotsParams{
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
				f.Is(e, management.ErrInvalidSearchQuery, httputil.NewInvalidField("query", "Must not exceed character limit."))
				f.Is(e, management.ErrInvalidCountryCode, httputil.NewInvalidField("country", "Must be a valid ISO-2 country code."))
				f.Is(e, management.ErrInvalidNorthEastLatitude, httputil.NewInvalidField("ne_lat", "Must be a valid latitude."))
				f.Is(e, management.ErrInvalidNorthEastLongitude, httputil.NewInvalidField("ne_lon", "Must be a valid longitude."))
				f.Is(e, management.ErrInvalidSouthWestLatitude, httputil.NewInvalidField("sw_lat", "Must be a valid latitude."))
				f.Is(e, management.ErrInvalidSouthWestLongitude, httputil.NewInvalidField("sw_lon", "Must be a valid longitude."))
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

	spot, err := h.service.CreateSpot(r.Context(), management.CreateSpotParams{
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
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotName, httputil.NewInvalidField("name", "Must be a non empty string."))
				f.Is(e, management.ErrInvalidCountryCode, httputil.NewInvalidField("country_code", "Must be a valid ISO-2 country code."))
				f.Is(e, management.ErrInvalidLocality, httputil.NewInvalidField("locality", "Must be a non empty string."))
				f.Is(e, management.ErrInvalidLatitude, httputil.NewInvalidField("latitude", "Must be a valid latitude."))
				f.Is(e, management.ErrInvalidLongitude, httputil.NewInvalidField("longitude", "Must be a valid longitude."))
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteCreated(w, r, toSpotResponse(spot))
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

	spot, err := h.service.UpdateSpot(r.Context(), management.UpdateSpotParams{
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
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, httputil.NewInvalidField(paramKeySpotID, "Must be a non empty string."))
				f.Is(e, management.ErrInvalidSpotName, httputil.NewInvalidField("name", "Must be a non empty string."))
				f.Is(e, management.ErrInvalidCountryCode, httputil.NewInvalidField("country_code", "Must be a valid ISO-2 country code."))
				f.Is(e, management.ErrInvalidLocality, httputil.NewInvalidField("locality", "Must be a non empty string."))
				f.Is(e, management.ErrInvalidLatitude, httputil.NewInvalidField("latitude", "Must be a valid latitude."))
				f.Is(e, management.ErrInvalidLongitude, httputil.NewInvalidField("longitude", "Must be a valid longitude."))
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, surf.ErrSpotNotFound) {
			httputil.WriteNotFoundError(w, r, "Such spot doesn't exist.")
			return
		}

		if errors.Is(err, surf.ErrEmptySpotUpdateEntry) {
			httputil.WriteValidationError(w, r, "Nothing to update.")
			return
		}

		httputil.WriteUnexpectedError(w, r, err)
		return
	}

	httputil.WriteOK(w, r, toSpotResponse(spot))
}

func (h *managementHandler) deleteSpot(w http.ResponseWriter, r *http.Request) {
	spotID := chi.URLParam(r, paramKeySpotID)

	if err := h.service.DeleteSpot(r.Context(), spotID); err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidSpotID, httputil.NewInvalidField(paramKeySpotID, "Must be a non empty string."))
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

	httputil.WriteNoContent(w, r)
}

func (h *managementHandler) location(w http.ResponseWriter, r *http.Request) {
	latitude, err := httputil.QueryParamFloat(r, "lat")
	if err != nil {
		httputil.WriteFieldError(w, r, httputil.NewInvalidField("lat", "Must be a valid latitude."))
		return
	}

	longitude, err := httputil.QueryParamFloat(r, "lon")
	if err != nil {
		httputil.WriteFieldError(w, r, httputil.NewInvalidField("lon", "Must be a valid longitude."))
		return
	}

	l, err := h.service.Location(r.Context(), geo.Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			f := httputil.NewInvalidFields()
			for _, e := range vErr.Errors() {
				f.Is(e, management.ErrInvalidLatitude, httputil.NewInvalidField("lat", "Must be a valid latitude."))
				f.Is(e, management.ErrInvalidLongitude, httputil.NewInvalidField("lon", "Must be a valid longitude."))
			}
			httputil.WriteFieldErrors(w, r, f)
			return
		}

		if errors.Is(err, geo.ErrLocationNotFound) {
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
