package commons

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const clientName = "empower-deposit-app"

type AuthData struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
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

	payloadb, err := base64.StdEncoding.DecodeString(authData.Payload)
	var authPayload AuthPayload
	if err := json.Unmarshal(payloadb, &authPayload); err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}

	if authPayload.Client != clientName {
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	pk, err := hex.DecodeString(authPayload.PubKey)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}
	pubKey := secp256k1.PubKey{Key: pk}

	sig, err := hex.DecodeString(authData.Signature)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "")
	}

	if ok := pubKey.VerifySignature([]byte(authData.Payload), sig); !ok {
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	return auth.UID(authPayload.PubKey), nil
}
