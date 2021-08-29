package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/management"
	"github.com/ztimes2/tolqin/internal/validation"
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

func (h *managementHandler) spot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName(paramKeySpotID)

	spot, err := h.service.Spot(id)
	if err != nil {
		if errors.Is(err, management.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusOK, fromManagementSpot(spot))
}

func (h *managementHandler) spots(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	limit, err := queryParamInt(r, "limit")
	if err != nil && !errors.Is(err, errEmptyParam) {
		writeError(w, r, http.StatusBadRequest, "Invalid limit.")
		return
	}

	offset, err := queryParamInt(r, "offset")
	if err != nil && !errors.Is(err, errEmptyParam) {
		writeError(w, r, http.StatusBadRequest, "Invalid offset.")
		return
	}

	countryCode := queryParam(r, "country")

	query := queryParam(r, "q")

	bounds, err := parseBounds(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid coordinates.")
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
		var vErr *validation.Error
		if errors.As(err, &vErr) {
			writeError(w, r, http.StatusBadRequest, vErr.Description())
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	resp := spotsResponse{
		Items: make([]spotResponse, len(spots)),
	}

	for i, s := range spots {
		resp.Items[i] = fromManagementSpot(s)
	}

	write(w, r, http.StatusOK, resp)
}

func (h *managementHandler) createSpot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var payload struct {
		Name        string  `json:"name"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Locality    string  `json:"locality"`
		CountryCode string  `json:"country_code"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid input.")
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
		var vErr *validation.Error
		if errors.As(err, &vErr) {
			writeError(w, r, http.StatusBadRequest, vErr.Description())
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusCreated, fromManagementSpot(spot))
}

func (h *managementHandler) updateSpot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	spotID := p.ByName(paramKeySpotID)

	var payload struct {
		Name        *string  `json:"name"`
		Latitude    *float64 `json:"latitude"`
		Longitude   *float64 `json:"longitude"`
		Locality    *string  `json:"locality"`
		CountryCode *string  `json:"country_code"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid input.")
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
		if errors.Is(err, management.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		if errors.Is(err, management.ErrNothingToUpdate) {
			writeError(w, r, http.StatusBadRequest, "Nothing to update.")
			return
		}
		var vErr *validation.Error
		if errors.As(err, &vErr) {
			writeError(w, r, http.StatusBadRequest, vErr.Description())
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusOK, fromManagementSpot(spot))
}

func (h *managementHandler) deleteSpot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	spotID := p.ByName(paramKeySpotID)

	if err := h.service.DeleteSpot(spotID); err != nil {
		if errors.Is(err, management.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusNoContent, nil)
}

func (h *managementHandler) location(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	latitude, err := queryParamFloat(r, "lat")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid latitude.")
		return
	}

	longitude, err := queryParamFloat(r, "lon")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid longitude.")
		return
	}

	l, err := h.service.Location(geo.Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		if errors.Is(err, management.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Location was not found.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	resp := locationResponse{
		Latitude:    l.Coordinates.Latitude,
		Longitude:   l.Coordinates.Longitude,
		Locality:    l.Locality,
		CountryCode: l.CountryCode,
	}

	write(w, r, http.StatusOK, resp)
}
