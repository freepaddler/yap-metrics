package ipmatcher

import (
	"net/http"
	"net/netip"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func IPMatchMiddleware(enabled bool, subnet netip.Prefix) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		ipMatch := func(w http.ResponseWriter, r *http.Request) {
			if enabled {
				requestIP := r.Header.Get("X-Real-IP")
				ip, err := netip.ParseAddr(requestIP)
				if err != nil {
					logger.Log().Debug().Err(err).Msgf("invalid X-Real-IP header `%s`", requestIP)
					w.WriteHeader(http.StatusForbidden)
					return
				}
				if !subnet.Contains(ip) {
					logger.Log().Debug().Err(err).Msgf("request rejected: ip %s does not match trusted subnet %s ", ip.String(), subnet.String())
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(ipMatch)
	}
}
