package xff

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

var privateMasks = func() []net.IPNet {
	masks := []net.IPNet{}
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7"} {
		_, net, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		masks = append(masks, *net)
	}
	return masks
}()

// IsPublicIP returns true if the given IP can be routed on the Internet
func IsPublicIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return false
	}
	for _, mask := range privateMasks {
		if mask.Contains(ip) {
			return false
		}
	}
	return true
}

// Parse parses the X-Forwarded-For Header and returns the IP address.
func Parse(ipList string) string {
	for _, ip := range strings.Split(ipList, ",") {
		ip = strings.TrimSpace(ip)
		if IP := net.ParseIP(ip); IP != nil && IsPublicIP(IP) {
			return ip
		}
	}
	return ""
}

// GetRemoteAddr parse the given request and resolve any X-Forwarded-* headers and return
// the resolved remote address.
func GetRemoteAddr(r *http.Request) string {
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
	remoteAddr := r.RemoteAddr
	if ip != "" && port != "" {
		remoteAddr = fmt.Sprintf("%s:%s", ip, port)
	} else {
		oip, oport, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if ip != "" {
				remoteAddr = fmt.Sprintf("%s:%s", ip, oport)
			} else if port != "" {
				remoteAddr = fmt.Sprintf("%s:%s", oip, port)
			}
		}
	}
	return remoteAddr
}

// XFF is a middleware to update RemoteAdd from X-Fowarded-* Headers.
func XFF(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = GetRemoteAddr(r)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
