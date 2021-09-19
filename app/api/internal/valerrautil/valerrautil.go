package valerrautil

import (
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
)

// IsCountry returns a valerra.Condition that checks if the given string is a
// valid ISO-2 country code.
func IsCountry(code string) valerra.Condition {
	return func() bool {
		return geo.IsCountry(code)
	}
}

// IsLatitude returns a valerra.Condition that checks if the given number is a
// valid latitude.
func IsLatitude(lat float64) valerra.Condition {
	return func() bool {
		return geo.IsLatitude(lat)
	}
}

// IsLongitude returns a valerra.Condition that checks if the given number is a
// valid longitude.
func IsLongitude(lon float64) valerra.Condition {
	return func() bool {
		return geo.IsLongitude(lon)
	}
}
