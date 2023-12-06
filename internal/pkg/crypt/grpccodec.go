package crypt

import (
	"crypto/rsa"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// KeyPairGRPCEncoder defines GPRC encoder, which allows to encode messages with RSA keypair.
// Codec should be registered on client and server.
// One keypair defines one way encryption: messages are encrypted by codec with defined public key
// and decrypted by codec with defined private key. To have two-way encryption, use 2 key pairs.
type KeyPairGRPCEncoder struct {
	pubKey  *rsa.PublicKey
	privKey *rsa.PrivateKey
	name    string
}

func NewKeyPairGRPCEncoder(name string, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) *KeyPairGRPCEncoder {
	logger.Log().Debug().Msgf("registering grpc codec `%s` pubKey: %t privkey: %t", name, pubKey != nil, privKey != nil)
	return &KeyPairGRPCEncoder{
		pubKey:  pubKey,
		privKey: privKey,
		name:    name,
	}
}

func (e *KeyPairGRPCEncoder) Name() string {
	return e.name
}

// Marshal tries to encrypt message.
// If public key is missing, then no encryption is made.
func (e *KeyPairGRPCEncoder) Marshal(v any) ([]byte, error) {
	logger.Log().Debug().Msg("encoding message")
	bytes, err := protojson.Marshal(v.(proto.Message))
	if err != nil {
		return nil, err
	}
	if e.pubKey != nil {
		return EncryptOAEP(e.pubKey, bytes)
	}
	return bytes, err
}

// Unmarshal tries to decrypt message.
// If private key is absent, message is considered as unencrypted.
// If decryption fails, message is considered as unencrypted.
// This behavior allows to control encryption only from client side
func (e *KeyPairGRPCEncoder) Unmarshal(data []byte, v any) error {
	logger.Log().Debug().Msg("decoding message")
	decrypted := data
	if e.privKey != nil {
		dec, err := DecryptOAEP(e.privKey, data)
		if err != nil {
			return err
		}
		decrypted = dec
	}
	return protojson.Unmarshal(decrypted, v.(proto.Message))
}
