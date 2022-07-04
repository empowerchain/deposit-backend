package commons

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encore.dev/beta/errs"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var defaultSigner = secp256k1.GenPrivKey()

func TestAuthHandler(t *testing.T) {
	tests := []struct {
		signerPrivKey *secp256k1.PrivKey
		pubKey        *secp256k1.PubKey
		client        string
		errorCode     errs.ErrCode
	}{
		{
			signerPrivKey: defaultSigner,
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        clientName,
			errorCode:     errs.OK,
		},
		{
			signerPrivKey: secp256k1.GenPrivKey(),
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        clientName,
			errorCode:     errs.Unauthenticated,
		},
		{
			signerPrivKey: defaultSigner,
			pubKey:        defaultSigner.PubKey().(*secp256k1.PubKey),
			client:        "invalidClientName",
			errorCode:     errs.Unauthenticated,
		},
	}

	for _, test := range tests {
		pubKeyHex := hex.EncodeToString(test.pubKey.Bytes())

		payload := fmt.Sprintf(`{"pubKey":"%s","client":"%s","timestamp":%d}`, pubKeyHex, test.client, time.Now().Unix())

		payloadSignatureB, err := test.signerPrivKey.Sign([]byte(payload))
		require.NoError(t, err)

		authData := fmt.Sprintf(`{
  "payload": %s,
  "signature": "%s"
}`, payload, hex.EncodeToString(payloadSignatureB))

		authDataB64 := base64.StdEncoding.EncodeToString([]byte(authData))
		uid, err := AuthHandler(context.Background(), authDataB64)

		if test.errorCode == errs.OK {
			require.NoError(t, err)
			require.Equal(t, string(uid), pubKeyHex)
		} else {
			require.Error(t, err)
			require.Equal(t, test.errorCode, err.(*errs.Error).Code)
		}
	}
}
