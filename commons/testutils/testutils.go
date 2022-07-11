package testutils

import (
	"context"
	"encoding/hex"
	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

var defaultSigner = secp256k1.GenPrivKey()

func GetAuthenticatedContext(uid string) context.Context {
	if uid == "" {
		uid = hex.EncodeToString(defaultSigner.PubKey().Bytes())
	}

	return auth.WithContext(context.Background(), auth.UID(uid), nil)
}

func GenerateKeys() (publicKey string, privateKey *secp256k1.PrivKey) {
	privateKey = secp256k1.GenPrivKey()
	publicKey = hex.EncodeToString(privateKey.PubKey().Bytes())
	return
}

func ClearDb(db string) error {
	_, err := sqldb.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", db))
	return err
}
