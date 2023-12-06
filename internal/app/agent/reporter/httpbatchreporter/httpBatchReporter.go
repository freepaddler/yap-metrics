package httpbatchreporter

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/compress"
	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/sign"
)

var (
	ErrBadResponse = errors.New("unexpected server response")
)

type Reporter struct {
	url       string
	client    http.Client
	key       string
	publicKey *rsa.PublicKey
	localIP   string
}

func (r Reporter) Close() error {
	return nil
}

func New(opts ...func(r *Reporter)) *Reporter {
	reporter := &Reporter{client: http.Client{}}
	for _, opt := range opts {
		opt(reporter)
	}
	return reporter
}

func WithAddress(a string) func(*Reporter) {
	return func(r *Reporter) {
		r.url = fmt.Sprintf("http://%s/updates/", a)
		r.localIP = getLocalIP(a)
	}
}

func WithHTTPTimeout(d time.Duration) func(*Reporter) {
	return func(r *Reporter) {
		r.client.Timeout = d
	}
}

func WithSignKey(k string) func(*Reporter) {
	return func(r *Reporter) {
		r.key = k
	}
}

func WithPublicKey(pk *rsa.PublicKey) func(*Reporter) {
	return func(r *Reporter) {
		r.publicKey = pk
	}
}

// getLocalIP gets local ip address used to reach server
func getLocalIP(addr string) string {
	parts := strings.Split(addr, ":")
	// port doesn't matter, no connection will be made
	conn, err := net.Dial("udp4", parts[0]+":1024")
	if err != nil {
		logger.Log().Err(err).Msgf("failed to get my ip with server address %s", addr)
		return ""
	}
	localAddress := conn.LocalAddr().(*net.UDPAddr)
	conn.Close()
	return localAddress.IP.String()
}

func (r Reporter) Send(m []models.Metrics) (err error) {
	log := logger.Log().With().Str("module", "httpBatchReporter").Logger()
	if len(m) == 0 {
		log.Info().Msg("skip sending: empty report")
		return
	}
	log.Debug().Msgf("sending %d metrics in batch", len(m))
	body, err := json.Marshal(m)
	if err != nil {
		log.Warn().Err(err).Msg("unable to marshal JSON batch")
		return
	}

	// calculate hash
	var HashSHA256 string
	if r.key != "" {
		HashSHA256 = sign.Get(body, r.key)
	}

	// encrypt body
	encBody := body
	if r.publicKey != nil {
		encBody, err = crypt.EncryptBody(&body, r.publicKey)
		if err != nil {
			log.Error().Err(err).Msg("unable to encrypt http request")
		}
	}

	// compress body
	reqBody, compressErr := compress.GzipBody(&encBody)

	req, err := http.NewRequest(http.MethodPost, r.url, reqBody)
	if err != nil {
		log.Error().Err(err).Msg("unable to create http request")
		return
	}

	// set X-Real-IP header
	if r.localIP != "" {
		req.Header.Set("X-Real-IP", r.localIP)
	}

	// set hash header
	if HashSHA256 != "" {
		req.Header.Set("HashSHA256", HashSHA256)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	if compressErr == nil {
		req.Header.Set("Content-Encoding", "gzip")
	}
	log.Debug().Msgf("sending metric %s", body)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to send metric %s", body)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// request failed
		log.Warn().Msgf("wrong http response status: %s", resp.Status)
		return ErrBadResponse
	}
	return
}
