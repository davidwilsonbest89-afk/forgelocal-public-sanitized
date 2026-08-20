// Command systemvault-doctor executes the native OS-vault acceptance subset.
// It never prints secret material; its JSON output is suitable for release evidence.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"forgelocal/internal/secrets"
)

type result struct {
	Service        string   `json:"service"`
	CreatedKey     bool     `json:"created_key"`
	ReadKey        bool     `json:"read_key"`
	RestartRead    bool     `json:"restart_read"`
	CreatedSecret  bool     `json:"created_secret"`
	ReadSecret     bool     `json:"read_secret"`
	Deleted        bool     `json:"deleted"`
	AbsentVerified bool     `json:"absent_verified"`
	ManualRequired []string `json:"manual_required"`
}

func main() {
	service := os.Getenv("FORGELOCAL_VAULT_SERVICE")
	if service == "" {
		service = fmt.Sprintf("ForgeLocal.Back01.Doctor.%d", time.Now().UnixNano())
	}
	id := "doctor.back01"
	secretID := "doctor.proxy"
	vault := secrets.NewSystemVault(service)
	key, err := secrets.NewKey()
	fatal(err)
	result := result{Service: service, ManualRequired: []string{
		"revoke or delete the OS-vault item outside ForgeLocal, then confirm Get returns ErrNotFound",
		"run under a restricted OS account/session and confirm Put fails without emitting secret material",
		"scan metadata.db, profile.json, logs and .flbackup after an integrated backup to confirm no secret value appears",
	}}
	fatal(vault.Put(id, key))
	result.CreatedKey = true
	readKey, err := vault.Get(id)
	fatal(err)
	result.ReadKey = len(readKey) == len(key) && subtle.ConstantTimeCompare(readKey, key) == 1
	if !result.ReadKey {
		fatal(errors.New("key round trip mismatch"))
	}

	// A new instance models a Core restart: it must retrieve the same OS-managed item.
	fresh := secrets.NewSystemVault(service)
	restartedKey, err := fresh.Get(id)
	fatal(err)
	result.RestartRead = len(restartedKey) == len(key) && subtle.ConstantTimeCompare(restartedKey, key) == 1
	if !result.RestartRead {
		fatal(errors.New("restart key round trip mismatch"))
	}

	value := []byte("vault-doctor-nonproduction-value")
	fatal(vault.PutSecret(secretID, value))
	result.CreatedSecret = true
	readValue, err := fresh.GetSecret(secretID)
	fatal(err)
	result.ReadSecret = len(readValue) == len(value) && subtle.ConstantTimeCompare(readValue, value) == 1
	if !result.ReadSecret {
		fatal(errors.New("secret round trip mismatch"))
	}

	fatal(fresh.Delete(id))
	fatal(fresh.DeleteSecret(secretID))
	result.Deleted = true
	_, keyErr := vault.Get(id)
	_, secretErr := vault.GetSecret(secretID)
	result.AbsentVerified = errors.Is(keyErr, secrets.ErrNotFound) && errors.Is(secretErr, secrets.ErrNotFound)
	if !result.AbsentVerified {
		fatal(errors.New("deleted vault values remained readable"))
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(true)
	fatal(enc.Encode(result))
}

func fatal(err error) {
	if err != nil {
		// The value may be supplied by an OS backend; retain only the operation class.
		fmt.Fprintln(os.Stderr, "system-vault matrix failed")
		os.Exit(1)
	}
}
