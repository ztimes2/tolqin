package valerrautil

import "github.com/ztimes2/tolqin/app/api/internal/geo"

func IsCountry(code string) func() bool {
	return func() bool {
		return geo.IsCountry(code)
	}
}

func IsLatitude(lat float64) func() bool {
	return func() bool {
		return geo.IsLatitude(lat)
	}
}

func IsLongitude(lon float64) func() bool {
	return func() bool {
		return geo.IsLongitude(lon)
	}
}
