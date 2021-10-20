package router

import (
	"errors"
	"strconv"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

var (
	errInvalidNorthEastLatitude  = errors.New("invalid north-east latitude")
	errInvalidNorthEastLongitude = errors.New("invalid north-east longitude")
	errInvalidSouthWestLatitude  = errors.New("invalid south-west latitude")
	errInvalidSouthWestLongitude = errors.New("invalid south-west longitude")
)

func parseBounds(neLat, neLon, swLat, swLon string) (*geo.Bounds, *valerra.Errors) {
	if neLat == "" && neLon == "" && swLat == "" && swLon == "" {
		return nil, nil
	}

	var (
		b    geo.Bounds
		err  error
		errs []error
	)

	b.NorthEast.Latitude, err = strconv.ParseFloat(neLat, 64)
	if err != nil {
		errs = append(errs, errInvalidNorthEastLatitude)
	}

	b.NorthEast.Longitude, err = strconv.ParseFloat(neLon, 64)
	if err != nil {
		errs = append(errs, errInvalidNorthEastLongitude)
	}

	b.SouthWest.Latitude, err = strconv.ParseFloat(swLat, 64)
	if err != nil {
		errs = append(errs, errInvalidSouthWestLatitude)
	}

	b.SouthWest.Longitude, err = strconv.ParseFloat(swLon, 64)
	if err != nil {
		errs = append(errs, errInvalidSouthWestLongitude)
	}

	if len(errs) == 0 {
		return &b, nil
	}

	return nil, valerra.NewErrors(errs...)
}
