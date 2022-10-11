package testutils

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	_ "encore.dev/appruntime/app/appinit"
	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const (
	AdminPubKey = "02714571c89d4a7626aa4fe07cf049b944f1b584781c4ec45662992180540b17f5"
)

var (
	adminDb   = sqldb.Named("admin")
	orgDB     = sqldb.Named("organization")
	depositDB = sqldb.Named("deposit")
	schemeDB  = sqldb.Named("scheme")
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
	if err := ClearDB(orgDB, "organization"); err != nil {
		panic(err)
	}
	if err := ClearDB(depositDB, "deposit", "voucher", "voucher_definition"); err != nil {
		panic(err)
	}
	if err := ClearDB(schemeDB, "scheme"); err != nil {
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

// EnsureExclusiveDatabaseAccess ensures the test runs with exclusive access to the database.
// No other tests that call EnsureExclusiveDatabaseAccess will have access during that time.
func EnsureExclusiveDatabaseAccess(t *testing.T) {
	ctx := context.Background()

	adminLockTx, err := adminDb.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	orgLockTx, err := orgDB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	depLockTx, err := depositDB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	schemeLockTx, err := schemeDB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		adminLockTx.Rollback()
		orgLockTx.Rollback()
		depLockTx.Rollback()
		schemeLockTx.Rollback()
	})

	if _, err := adminLockTx.Exec(ctx, "SELECT pg_advisory_lock(1)"); err != nil {
		t.Fatal(err)
	}

	if _, err := orgLockTx.Exec(ctx, "SELECT pg_advisory_lock(1)"); err != nil {
		t.Fatal(err)
	}

	if _, err := depLockTx.Exec(ctx, "SELECT pg_advisory_lock(1)"); err != nil {
		t.Fatal(err)
	}

	if _, err := schemeLockTx.Exec(ctx, "SELECT pg_advisory_lock(1)"); err != nil {
		t.Fatal(err)
	}
}
