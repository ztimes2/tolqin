package router

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/surfing"
	"github.com/ztimes2/tolqin/internal/validation"
)

type surfingService interface {
	Spot(id string) (surfing.Spot, error)
	Spots(surfing.SpotsParams) ([]surfing.Spot, error)
}

func fromSurfingSpot(s surfing.Spot) spotResponse {
	return spotResponse{
		ID:          s.ID,
		Name:        s.Name,
		Latitude:    s.Location.Coordinates.Latitude,
		Longitude:   s.Location.Coordinates.Longitude,
		Locality:    s.Location.Locality,
		CountryCode: s.Location.CountryCode,
	}
}

type surfingHandler struct {
	service surfingService
}

func newSurfingHandler(s surfingService) *surfingHandler {
	return &surfingHandler{
		service: s,
	}
}

func (h *surfingHandler) spot(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

	write(w, r, http.StatusOK, fromSurfingSpot(spot))
}

func (h *surfingHandler) spots(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		resp.Items[i] = fromSurfingSpot(s)
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
