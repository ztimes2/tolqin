package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/validation"
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
		if errors.Is(err, surfer.ErrNotFound) {
			writeError(w, r, http.StatusNotFound, "Such spot doesn't exist.")
			return
		}
		writeUnexpectedError(w, r, err)
		return
	}

	write(w, r, http.StatusOK, fromSurferSpot(spot))
}

func (h *surferHandler) spots(w http.ResponseWriter, r *http.Request) {
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

	spots, err := h.service.Spots(surfer.SpotsParams{
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
		resp.Items[i] = fromSurferSpot(s)
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
