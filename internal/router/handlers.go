package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/surfing"
	"github.com/ztimes2/tolqin/internal/validation"
)

type spotResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Locality    string  `json:"locality"`
	CountryCode string  `json:"country_code"`
}

func toSpotResponse(s surfing.Spot) spotResponse {
	return spotResponse{
		ID:          s.ID,
		Name:        s.Name,
		Latitude:    s.Location.Coordinates.Latitude,
		Longitude:   s.Location.Coordinates.Longitude,
		Locality:    s.Location.Locality,
		CountryCode: s.Location.CountryCode,
	}
}

type spotsResponse struct {
	Items []spotResponse `json:"items"`
}

type handler struct {
	service service
}

func newHandler(s service) *handler {
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
	limit, err := queryParamInt(r, "limit")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid limit.")
		return
	}

	offset, err := queryParamInt(r, "offset")
	if err != nil {
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

	spots, err := h.service.Spots(surfing.SpotsParams{
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
		resp.Items[i] = toSpotResponse(s)
	}

	write(w, r, http.StatusOK, resp)
}

func parseBounds(r *http.Request) (*geo.Bounds, error) {
	neLat := queryParam(r, "ne_lat")
	neLon := queryParam(r, "ne_lon")
	swLat := queryParam(r, "sw_lat")
	swLon := queryParam(r, "sw_lon")

	if neLat == "" && neLon == "" && swLat == "" && swLon == "" {
		return nil, nil
	}

	var (
		b   geo.Bounds
		err error
	)

	b.NorthEast.Latitude, err = strconv.ParseFloat(neLat, 64)
	if err != nil {
		return nil, errors.New("invalid north-east latitude")
	}
	b.NorthEast.Longitude, err = strconv.ParseFloat(neLon, 64)
	if err != nil {
		return nil, errors.New("invalid north-east longitude")
	}
	b.SouthWest.Latitude, err = strconv.ParseFloat(swLat, 64)
	if err != nil {
		return nil, errors.New("invalid south-west latitude")
	}
	b.SouthWest.Longitude, err = strconv.ParseFloat(swLon, 64)
	if err != nil {
		return nil, errors.New("invalid south-west longitude")
	}

	return &b, nil
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
		Name: payload.Name,
		Coordinates: geo.Coordinates{
			Latitude:  payload.Latitude,
			Longitude: payload.Longitude,
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

	params := surfing.UpdateSpotParams{
		ID:   spotID,
		Name: payload.Name,
	}
	if payload.Latitude != nil || payload.Longitude != nil {
		c := &geo.Coordinates{}
		if payload.Latitude != nil {
			c.Latitude = *payload.Latitude
		}
		if payload.Longitude != nil {
			c.Longitude = *payload.Longitude
		}
		params.Coordinates = c
	}

	spot, err := h.service.UpdateSpot(params)
	if err != nil {
		if errors.Is(err, surfing.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		if errors.Is(err, surfing.ErrNothingToUpdate) {
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
