package commons

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const clientName = "empower-deposit-app"

type AuthData struct {
	Payload   AuthPayload `json:"payload"`
	Signature string      `json:"signature"`
}

type AuthPayload struct {
	PubKey    string `json:"pubKey"`
	Client    string `json:"client"`
	Timestamp int64  `json:"timestamp"`
}

//encore:authhandler
func AuthHandler(_ context.Context, token string) (auth.UID, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}
	var authData AuthData
	if err := json.Unmarshal(b, &authData); err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}

	if authData.Payload.Client != clientName {
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	pk, err := hex.DecodeString(authData.Payload.PubKey)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}
	pubKey := secp256k1.PubKey{Key: pk}

	sig, err := hex.DecodeString(authData.Signature)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}

	payload := fmt.Sprintf(`{"pubKey":"%s","client":"%s","timestamp":%d}`, authData.Payload.PubKey, authData.Payload.Client, authData.Payload.Timestamp)
	if ok := pubKey.VerifySignature([]byte(payload), sig); !ok {
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	return auth.UID(authData.Payload.PubKey), nil
}
