package tmp

import (
	"context"
	"encoding/hex"
	"encore.app/commons"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

//encore:api auth method=GET
func AuthTest(_ context.Context) error {
	return nil
}

//encore:api public method=GET
func StatusTest(_ context.Context) error {
	return nil
}

type GenerateKeyResponse struct {
	PublicKey  string
	PrivateKey string
	Token      string
}

//encore:api public method=GET
func GenerateKey(_ context.Context) (*GenerateKeyResponse, error) {
	privateKey := secp256k1.GenPrivKey()
	publicKey := privateKey.PubKey().(*secp256k1.PubKey)

	token, pubKeyHex, err := commons.GetToken(privateKey, publicKey, commons.ClientName)
	if err != nil {
		return nil, err
	}

	return &GenerateKeyResponse{
		PublicKey:  pubKeyHex,
		PrivateKey: hex.EncodeToString(privateKey.Bytes()),
		Token:      token,
	}, nil
}
