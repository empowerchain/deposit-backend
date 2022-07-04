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

func TestAuthHandlerPreset(t *testing.T) {
	uid, err := AuthHandler(context.Background(), "ewogICJwYXlsb2FkIjogImV3b2dJQ0p3ZFdKTFpYa2lPaUFpTURKbU5EUTJOVEF5WkRCaE1HVXlPR1prWVRReU1EaGlabUZqTlRNNFpHWTROakl4Tm1ReU1ETTBZVGhqWldJeU1qWTROMk14TTJJNE1XUTBaalZsWkRNMElpd0tJQ0FpWTJ4cFpXNTBJam9pWlcxd2IzZGxjaTFrWlhCdmMybDBMV0Z3Y0NJc0NpQWdJblJwYldWemRHRnRjQ0k2TVRZMU5qazBNekl3TndwOSIsCiAgInNpZ25hdHVyZSI6ICI4OWJhNDA5ZjNmZDAxYzE1YzAwNGMzNjA4NjUzYzRkMjFkZmU4NjY4ZTc0N2M4ZDljNmM1N2M1ZDQzOTY3NjQ5NzUxYzVjZGQyZDE3YzRjMTVhY2E2MTE1YTVhYzZiMDIzZDcxYWY2NjJmMThkMGNlOWU2MzdmZTE5NDI0NmY1ZSIKfQ==")
	require.NoError(t, err)
	require.Equal(t, "02f446502d0a0e28fda4208bfac538df86216d2034a8ceb22687c13b81d4f5ed34", string(uid))
}

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

		payloadStr := fmt.Sprintf(`{
  "pubKey": "%s",
  "client":"%s",
  "timestamp":%d
}`, pubKeyHex, test.client, time.Now().Unix())
		payload := base64.StdEncoding.EncodeToString([]byte(payloadStr))

		payloadSignatureB, err := test.signerPrivKey.Sign([]byte(payload))
		require.NoError(t, err)

		authData := fmt.Sprintf(`{
  "payload": "%s",
  "signature": "%s"
}`, payload, hex.EncodeToString(payloadSignatureB))

		authDataB64 := base64.StdEncoding.EncodeToString([]byte(authData))
		fmt.Println(authDataB64)
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
