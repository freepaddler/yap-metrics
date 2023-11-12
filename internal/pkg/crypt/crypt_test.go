package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	privateValid = `
-----BEGIN PRIVATE KEY-----
MGMCAQACEQDMyxsFzwkld564GOeSsZRhAgMBAAECEEqT5EFXReNoZyy5Qd/XcAEC
CQD1pqF4dvAZwQIJANVr0on6NeKhAgkAx0fsijuCwIECCQDEKJIel4cngQIIMCeA
Sl6DKjY=
-----END PRIVATE KEY-----
`
	privateBadHeader = `
-----BEGIN SOME PRIVATE KEY-----
MGMCAQACEQDMyxsFzwkld564GOeSsZRhAgMBAAECEEqT5EFXReNoZyy5Qd/XcAEC
CQD1pqF4dvAZwQIJANVr0on6NeKhAgkAx0fsijuCwIECCQDEKJIel4cngQIIMCeA
Sl6DKjY=
-----END SOME PRIVATE KEY-----
`
	privateInvalid = `
-----BEGIN PRIVATE KEY-----
MGMCAQACEQDMyxsFzwkld564GOeSsZRhAgMBAAECEEqT5EFXReNoZyy5Qd/XcAEC
CQD1pqF4dvAZwDDJANVr0on6NeKhAgkAx0fsijuCwIECCQDEKJIel4cngQIIMCeA
Sl6DKjY=
-----END PRIVATE KEY-----
`
	publicValid = `
-----BEGIN PUBLIC KEY-----
MBgCEQDMyxsFzwkld564GOeSsZRhAgMBAAE=
-----END PUBLIC KEY-----
`
	publicBadHeader = `
-----BEGIN SOME PUBLIC KEY-----
MBgCEQDMyxsFzwkld564GOeSsZRhAgMBAAE=
-----END SOME PUBLIC KEY-----
`
	publicInvalid = `
some 
fake strings 
there
`
)

func Test_EncryptDecrypt(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	payload := make([]byte, 3000)
	rand.Read(payload)
	encrypted, err := EncryptOAEP(&key.PublicKey, payload)
	require.NoError(t, err, "No encryption errors expected")
	require.NotEqual(t, encrypted, payload, "Expect payload to be encrypted")
	decrypted, err := DecryptOAEP(key, encrypted)
	require.NoError(t, err, "No decryption errors expected")
	require.Equal(t, decrypted, payload, "Expect decrypted to match payload")

}

func TestReadPublicKey(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr error
	}{
		{
			name:    "invalid",
			key:     []byte(publicInvalid),
			wantErr: ErrInvalidPEM,
		},
		{
			name:    "invalid header",
			key:     []byte(publicBadHeader),
			wantErr: ErrInvalidPEM,
		},
		{
			name: "valid",
			key:  []byte(publicValid),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewReader(tt.key)
			_, err := ReadPublicKey(buf)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, ErrInvalidPEM)
			}
		})
	}
}

func TestReadPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr error
	}{
		{
			name:    "invalid",
			key:     []byte(privateInvalid),
			wantErr: ErrInvalidPEM,
		},
		{
			name:    "invalid header",
			key:     []byte(privateBadHeader),
			wantErr: ErrInvalidPEM,
		},
		{
			name: "valid",
			key:  []byte(privateValid),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewReader(tt.key)
			priv, err := ReadPrivateKey(buf)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, ErrInvalidPEM)
			} else {
				require.NoError(t, priv.Validate())
			}
		})
	}
}

func TestEncryptBody(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	payload := make([]byte, 3000)
	body, err := EncryptOAEP(&key.PublicKey, payload)
	require.NoError(t, err, "No encryption errors expected")
	require.NotEqual(t, body, payload, "Expect payload to be encrypted")
}

func TestDecryptMiddleware(t *testing.T) {
	payload := make([]byte, 3000)
	rand.Read(payload)
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	testHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "Expect body to be read")
		defer r.Body.Close()
		if string(body) == string(payload) {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	})

	tests := []struct {
		name     string
		priv     *rsa.PrivateKey
		pub      *rsa.PublicKey
		wantCode int
	}{
		{
			name:     "encrypted flow",
			priv:     key,
			pub:      &key.PublicKey,
			wantCode: http.StatusOK,
		},
		{
			name:     "server with key, client without key",
			priv:     key,
			pub:      nil,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "server without key, client with key",
			priv:     nil,
			pub:      &key.PublicKey,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "no encryption",
			priv:     nil,
			pub:      nil,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := payload
			if tt.pub != nil {
				body, err = EncryptBody(&body, tt.pub)
				require.NoError(t, err, "Expect no error on encryption")
			}
			reqBody := bytes.NewReader(body)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			// middlware
			w := httptest.NewRecorder()
			mw := DecryptMiddleware(tt.priv)
			handler := mw(testHandler)
			handler.ServeHTTP(w, req)

			require.Equal(t, tt.wantCode, w.Code)
		})
	}
}
