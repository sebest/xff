package xff

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var (
	privateIPRE = regexp.MustCompile("(^127.0.0.1)|(^10.)|(^172.1[6-9].)|(^172.2[0-9].)|(^172.3[0-1].)|(^192.168.)")
	ipRE        = regexp.MustCompile("^[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}$")
)

// Parse parses the X-Forwarded-For Header and returns the IP address.
func Parse(ipAddress string) string {
	if ipRE.MatchString(ipAddress) && !privateIPRE.MatchString(ipAddress) {
		return ipAddress
	}

	parts := strings.Split(ipAddress, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
		if privateIPRE.MatchString(parts[i]) {
			continue
		} else if ipRE.MatchString(parts[i]) {
			return parts[i]
		}
	}
	return ""
}

func parseXFP(port string) string {
	return port
}

// XFF is a middleware to update RemoteAdd from X-Fowarded-* Headers.
func XFF(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		xff := r.Header.Get("X-Forwarded-For")
		xfp := r.Header.Get("X-Forwarded-Port")
		var ip string
		if xff != "" {
			ip = Parse(xff)
		}
		var port string
		if xfp != "" {
			port = Parse(xfp)
		}
		if ip != "" && port != "" {
			r.RemoteAddr = fmt.Sprintf("%s:%s", ip, port)
		} else {
			oip, oport, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				if ip != "" {
					r.RemoteAddr = fmt.Sprintf("%s:%s", ip, oport)

				} else if port != "" {
					r.RemoteAddr = fmt.Sprintf("%s:%s", oip, port)

				}
			}
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
