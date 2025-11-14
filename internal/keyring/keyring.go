package keyring

import (
	"github.com/zalando/go-keyring"
)

// Set stores a secret in the keyring
func Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

// Get retrieves a secret from the keyring
func Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}
