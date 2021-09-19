package valerrautil

import "github.com/ztimes2/tolqin/app/api/internal/geo"

// IsCountry returns a condition function that checks if the given string is a
// valid ISO-2 country code.
func IsCountry(code string) func() bool {
	return func() bool {
		return geo.IsCountry(code)
	}
}

// IsLatitude returns a condition function that checks if the given number is a
// valid latitude.
func IsLatitude(lat float64) func() bool {
	return func() bool {
		return geo.IsLatitude(lat)
	}
}

// IsLongitude returns a condition function that checks if the given number is a
// valid longitude.
func IsLongitude(lon float64) func() bool {
	return func() bool {
		return geo.IsLongitude(lon)
	}
}
