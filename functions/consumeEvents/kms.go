package main

import (
	"os"

	"github.com/dwtechnologies/kmsdecrypt"
)

// decryptPassword will decrypt the password string provided by password.
// And set the encrypted password to the pass variable.
// Returns error.
func decryptPassword(password string) error {
	decrypter, err := kmsdecrypt.New(os.Getenv("AWS_DEFAULT_REGION"))
	if err != nil {
		return err
	}

	decrypted, err := decrypter.DecryptString(password)
	if err != nil {
		return err
	}

	pass = decrypted
	return nil
}
