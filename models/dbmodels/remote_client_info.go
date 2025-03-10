package dbmodels

import (
	"github.com/strongo/validation"
	"strings"
)

// RemoteClientInfo a struct to store data about remote client (e.g. IP address)
type RemoteClientInfo struct {
	HostOrApp    string    `json:"hostOrApp" firestore:"hostOrApp"`
	RemoteAddr   string    `json:"remoteAddr,omitempty" firestore:"remoteAddr,omitempty"` // Can be enpty in case of messenger
	ForwardedFor string    `json:"forwardedFor,omitempty" firestore:"forwardedFor,omitempty"`
	GeoCountry   string    `json:"geoCountry,omitempty" firestore:"geoCountry,omitempty"`
	GeoRegion    string    `json:"geoRegion,omitempty" firestore:"geoRegion,omitempty"`
	GeoCity      string    `json:"geoCity,omitempty" firestore:"geoCity,omitempty"`
	GeoCityPoint *GeoPoint `json:"geoCityPoint,omitempty" firestore:"geoCityPoint,omitempty"`
}

func (v RemoteClientInfo) Validate() error {
	if strings.TrimSpace(v.HostOrApp) == "" {
		return validation.NewErrRecordIsMissingRequiredField("hostOrApp")
	}
	if strings.TrimSpace(v.RemoteAddr) == "" &&
		!strings.Contains(v.HostOrApp, "@") { // @ is used to separate platform and bot userID
		return validation.NewErrRecordIsMissingRequiredField("remoteAddr")
	}
	if v.GeoCityPoint != nil && !v.GeoCityPoint.Valid() {
		return validation.NewErrBadRecordFieldValue("geoCityPoint", "invalid GeoPoint")
	}
	return nil
}

// GeoPoint represents a location as latitude/longitude in degrees.
type GeoPoint struct {
	Lat, Lng float64
}

// Valid returns whether a GeoPoint is within [-90, 90] latitude and [-180, 180] longitude.
func (g GeoPoint) Valid() bool {
	return -90 <= g.Lat && g.Lat <= 90 && -180 <= g.Lng && g.Lng <= 180
}
