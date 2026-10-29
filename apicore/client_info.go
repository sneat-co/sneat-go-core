package apicore

import (
	"net/http"
	"net/url"

	"github.com/sneat-co/sneat-go-core/models/dbmodels"
)

func GetRemoteClientInfo(r *http.Request) dbmodels.RemoteClientInfo {
	header := r.Header
	client := dbmodels.RemoteClientInfo{
		HostOrApp:    r.Host,
		Origin:       requestOrigin(header),
		RemoteAddr:   header.Get("CF-Connecting-IP"),
		ForwardedFor: header.Get("X-Forwarded-For"),
		GeoCountry:   header.Get("CF-IPCountry"),
		//GeoCountry: header.Get("X-Appengine-Country"),
		//GeoRegion: header.Get("X-Appengine-Region"),
		//GeoCity:   header.Get("X-Appengine-City"),
	}
	if client.RemoteAddr == "" {
		client.RemoteAddr = r.RemoteAddr
	}
	if client.ForwardedFor == client.RemoteAddr { // This is probably neven going to happen?
		client.ForwardedFor = ""
	}
	return client
}

// requestOrigin returns the scheme and host of the website that sent the
// request: the Origin header, else the Referer stripped of path and query.
func requestOrigin(header http.Header) string {
	for _, raw := range []string{header.Get("Origin"), header.Get("Referer")} {
		if u, err := url.Parse(raw); err == nil && u.Scheme != "" && u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
	}
	return ""
}
