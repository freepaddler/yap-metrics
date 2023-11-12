// Package crypt provides ability to encrypt client-server communication.
// The encryption is one-way client to server and implemented with RSA keypair.
package crypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	privateType = "PRIVATE KEY"
	publicType  = "PUBLIC KEY"
)

var (
	ErrDecrypt    = errors.New("decryption error")
	ErrEncrypt    = errors.New("encryption error")
	ErrInvalidPEM = errors.New("invalid PEM file")
	ErrGenerate   = errors.New("invalid PEM file")
	useHash       = sha256.New()
	useRand       = rand.Reader
)

// WritePair generates and writes RSA key pair into io.Writers
func WritePair(pubOut, privOut io.Writer, n int) error {
	key, err := rsa.GenerateKey(rand.Reader, n)
	if err != nil {
		return fmt.Errorf("%w (%w)", ErrGenerate, err)
	}

	err = pem.Encode(privOut, &pem.Block{
		Type:  privateType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return fmt.Errorf("%w (%w)", ErrGenerate, err)
	}
	err = pem.Encode(pubOut, &pem.Block{
		Type:  publicType,
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	})
	if err != nil {
		return fmt.Errorf("%w (%w)", ErrGenerate, err)
	}
	return nil
}

// ReadPublicKey reads public key in PEM format
func ReadPublicKey(in io.Reader) (*rsa.PublicKey, error) {
	data, _ := io.ReadAll(in)
	block, _ := pem.Decode(data)
	if block == nil || block.Type != publicType {
		return nil, ErrInvalidPEM
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPEM
	}
	return pub, nil
}

// ReadPrivateKey reads private key in PEM format
func ReadPrivateKey(in io.Reader) (*rsa.PrivateKey, error) {
	data, _ := io.ReadAll(in)
	block, _ := pem.Decode(data)
	if block == nil || block.Type != privateType {
		return nil, ErrInvalidPEM
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrInvalidPEM
	}
	return priv, nil
}

// EncryptOAEP allows to make rsa encryption for long messages, splitting encryption into blocks.
func EncryptOAEP(pub *rsa.PublicKey, msg []byte) ([]byte, error) {
	msgSize := len(msg)
	blockSize := pub.Size() - 2*useHash.Size() - 2
	var encryptedBytes []byte
	for startIdx := 0; startIdx < msgSize; startIdx += blockSize {
		endIdx := startIdx + blockSize
		if endIdx > msgSize {
			endIdx = msgSize
		}
		encryptedBlockBytes, err := rsa.EncryptOAEP(useHash, useRand, pub, msg[startIdx:endIdx], nil)
		if err != nil {
			return nil, fmt.Errorf("%w (%w)", ErrEncrypt, err)
		}
		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}
	return encryptedBytes, nil
}

// DecryptOAEP allows to decrypt long messages encrypted with EncryptOAEP
func DecryptOAEP(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	msgSize := len(ciphertext)
	blockSize := priv.PublicKey.Size()
	var decryptedBytes []byte
	for startIdx := 0; startIdx < msgSize; startIdx += blockSize {
		endIdx := startIdx + blockSize
		if endIdx > msgSize {
			endIdx = msgSize
		}
		decryptedBlockBytes, err := rsa.DecryptOAEP(useHash, nil, priv, ciphertext[startIdx:endIdx], nil)
		if err != nil {
			return nil, fmt.Errorf("%w (%w)", ErrDecrypt, err)
		}
		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}
	return decryptedBytes, nil
}

// EncryptBody tries to encrypt response body.
// Returns unencrypted body and error if failed.
func EncryptBody(body *[]byte, pubKey *rsa.PublicKey) ([]byte, error) {
	encrypted, err := EncryptOAEP(pubKey, *body)
	if err != nil {
		return *body, fmt.Errorf("%w (%w)", ErrEncrypt, err)
	}
	logger.Log().Debug().Msg("body encrypted")
	return encrypted, nil

}

func DecryptMiddleware(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		decMW := func(w http.ResponseWriter, r *http.Request) {
			if privateKey != nil {
				reqBody, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Log().Warn().Err(err).Msg("failed to read request body")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				defer r.Body.Close()
				decrypted, err := DecryptOAEP(privateKey, reqBody)
				if err != nil {
					logger.Log().Warn().Err(err).Msg("failed to decrypt request body")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				plainBody := bytes.NewReader(decrypted)
				r.Body = io.NopCloser(plainBody)
				logger.Log().Debug().Msg("body decrypted")
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(decMW)
	}
}

//
//func DecryptBody(body io.ReadCloser, privKey *rsa.PrivateKey) (io.ReadCloser, error) {
//	reqBody, err := io.ReadAll(body)
//	if err != nil {
//		return nil, ErrRead
//	}
//	defer body.Close()
//	decrypted, err := rsa.DecryptOAEP(sha256.New(), nil, privKey, reqBody, nil)
//	if err != nil {
//		return nil, ErrDecrypt
//	}
//	var buf bytes.Buffer
//	buf.Write(decrypted)
//	return io.NopCloser(&buf), nil
//}
