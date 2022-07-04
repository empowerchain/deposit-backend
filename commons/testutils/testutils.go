package testutils

import (
	"context"
	"encoding/hex"
	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

var defaultSigner = secp256k1.GenPrivKey()

func GetAuthenticatedContext() context.Context {
	pubKeyHex := hex.EncodeToString(defaultSigner.PubKey().Bytes())
	return auth.WithContext(context.Background(), auth.UID(pubKeyHex), nil)
}

func ClearDb() error {
	_, err := sqldb.Exec(context.Background(), "TRUNCATE scheme")
	return err
}
