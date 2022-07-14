package testutils

import (
	"context"
	"encoding/hex"
	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const (
	AdminPubKey = "02714571c89d4a7626aa4fe07cf049b944f1b584781c4ec45662992180540b17f5"
)

var (
	voucherDB = sqldb.Named("voucher")
	orgDB     = sqldb.Named("organization")
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

func ClearAllDBs() {
	if err := ClearDB(voucherDB, "voucher_definition", "voucher"); err != nil {
		panic(err)
	}
	if err := ClearDB(orgDB, "organization"); err != nil {
		panic(err)
	}
}

func ClearDB(db *sqldb.Database, tables ...string) error {
	for _, table := range tables {
		if _, err := db.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return err
		}
	}

	return nil
}
