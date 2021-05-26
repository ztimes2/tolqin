package router

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/ztimes2/tolqin/backend/internal/surfing"
)

type spotResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func toSpotResponse(s surfing.Spot) spotResponse {
	return spotResponse{
		ID:        s.ID,
		Name:      s.Name,
		Latitude:  s.Latitude,
		Longitude: s.Longitude,
	}
}

type handler struct {
	service *surfing.Service
}

func newHandler(s *surfing.Service) *handler {
	return &handler{
		service: s,
	}
}

func (h *handler) spot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName(paramKeySpotID)

	spot, err := h.service.Spot(id)
	if err != nil {
		if errors.Is(err, surfing.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusOK, toSpotResponse(spot))
}

func (h *handler) spots(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	spots, err := h.service.Spots()
	if err != nil {
		if errors.Is(err, surfing.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	var items []interface{}
	for _, s := range spots {
		items = append(items, toSpotResponse(s))
	}

	write(w, r, http.StatusOK, toListResponse(items))
}

func (h *handler) createSpot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var payload struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid input.")
		return
	}

	spot, err := h.service.CreateSpot(surfing.CreateSpotParams{
		Name:      payload.Name,
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
	})
	if err != nil {
		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			writeError(w, r, http.StatusBadRequest, humanizeValidationErrors(vErr))
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusCreated, toSpotResponse(spot))
}

func (h *handler) updateSpot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	spotID := p.ByName(paramKeySpotID)

	var payload struct {
		Name      *string  `json:"name"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid input.")
		return
	}

	spot, err := h.service.UpdateSpot(surfing.UpdateSpotParams{
		ID:        spotID,
		Name:      payload.Name,
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
	})
	if err != nil {
		if errors.Is(err, surfing.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		if errors.Is(err, surfing.ErrNothingToUpdate) {
			writeError(w, r, http.StatusBadRequest, "Nothing to update.")
			return
		}
		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			writeError(w, r, http.StatusBadRequest, humanizeValidationErrors(vErr))
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusOK, toSpotResponse(spot))
}

func (h *handler) deleteSpot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	spotID := p.ByName(paramKeySpotID)

	if err := h.service.DeleteSpot(spotID); err != nil {
		if errors.Is(err, surfing.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusNoContent, nil)
}
