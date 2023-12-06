package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
)

func TestKeyPairGRPCEncoder(t *testing.T) {
	privKey1, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	privkey2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	req := pb.MetricsBatch{
		Metrics: []*pb.Metric{
			{
				Id:    "c1",
				Type:  pb.Metric_COUNTER,
				Delta: 10,
			},
			{
				Id:    "g1",
				Type:  pb.Metric_GAUGE,
				Value: -0.117,
			},
		},
	}
	tests := []struct {
		name    string
		pubKey  *rsa.PublicKey
		privKey *rsa.PrivateKey
		wantErr bool
	}{
		{
			name: "no keys",
		},
		{
			name:    "valid pair",
			pubKey:  &privKey1.PublicKey,
			privKey: privKey1,
		},
		{
			name:    "invalid pair",
			pubKey:  &privKey1.PublicKey,
			privKey: privkey2,
			wantErr: true,
		},
		{
			name:    "no public key",
			privKey: privkey2,
			wantErr: true,
		},
		{
			name:    "no private key",
			pubKey:  &privkey2.PublicKey,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := &KeyPairGRPCEncoder{privKey: tt.privKey, pubKey: tt.pubKey}
			b, err := encoder.Marshal(&req)
			require.NoError(t, err, "no error expected on marshal")
			res := new(pb.MetricsBatch)
			err = encoder.Unmarshal(b, res)
			if tt.wantErr {
				require.Error(t, err, "error expected on unmarshall")
			} else {
				require.NoError(t, err, "no error expected on unmarshall")
				require.Equalf(t, &req, res, "expect equal request and response: \n%+v\n%+v\n", &req, res)
			}
		})
	}
}

func TestNewKeyPairGRPCEncoder(t *testing.T) {
	name := "somename"
	e := NewKeyPairGRPCEncoder(name, nil, nil)
	got := e.Name()
	assert.Equal(t, name, got)
}
