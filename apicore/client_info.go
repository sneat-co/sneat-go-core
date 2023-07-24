package apicore

import (
	"github.com/sneat-co/sneat-go/src/models/dbmodels"
	"net/http"
)

func GetRemoteClientInfo(r *http.Request) dbmodels.RemoteClientInfo {
	header := r.Header
	client := dbmodels.RemoteClientInfo{
		HostOrApp:    r.Host,
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
