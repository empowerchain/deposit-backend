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
	"time"
)

const ClientName = "empower-deposit-app"

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
		return "", errs.WrapCode(err, errs.Unauthenticated, "failed to decode token from base64")
	}
	var authData AuthData
	if err := json.Unmarshal(b, &authData); err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "failed to unmarshal token from JSON")
	}

	payloadb, err := base64.StdEncoding.DecodeString(authData.Payload)
	var authPayload AuthPayload
	if err := json.Unmarshal(payloadb, &authPayload); err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "failed to decode payload from base64")
	}

	if authPayload.Client != ClientName {
		return "", &errs.Error{
			Code: errs.Unauthenticated,
		}
	}

	pk, err := hex.DecodeString(authPayload.PubKey)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "failed to decode pubKey from hex")
	}
	pubKey := secp256k1.PubKey{Key: pk}

	sig, err := hex.DecodeString(authData.Signature)
	if err != nil {
		return "", errs.WrapCode(err, errs.Unauthenticated, "failed to decode signature from hex")
	}

	if ok := pubKey.VerifySignature([]byte(authData.Payload), sig); !ok {
		return "", &errs.Error{
			Message: "failed to verify signature",
			Code:    errs.Unauthenticated,
		}
	}

	return auth.UID(authPayload.PubKey), nil
}

// FYI, kept here instead of test because it is used by tmp.GenerateKey
func GetToken(signerPrivKey *secp256k1.PrivKey, pubKey *secp256k1.PubKey, client string) (string, string, error) {
	pubKeyHex := hex.EncodeToString(pubKey.Bytes())

	payloadStr := fmt.Sprintf(`{
  "pubKey": "%s",
  "client":"%s",
  "timestamp":%d
}`, pubKeyHex, client, time.Now().Unix())
	payload := base64.StdEncoding.EncodeToString([]byte(payloadStr))

	payloadSignatureB, err := signerPrivKey.Sign([]byte(payload))
	if err != nil {
		return "", "", err
	}

	authData := fmt.Sprintf(`{
  "payload": "%s",
  "signature": "%s"
}`, payload, hex.EncodeToString(payloadSignatureB))

	authDataB64 := base64.StdEncoding.EncodeToString([]byte(authData))

	return authDataB64, pubKeyHex, nil
}
