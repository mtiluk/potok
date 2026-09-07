package secrets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const service = "potok"
const APIKey = "api-key"

var (
	ErrNotFound    = errors.New("secrets: no credential stored")
	ErrUnavailable = errors.New("secrets: no keyring available")
)

func VaultKeyName(vault string) string {
	return "vault:" + vault
}

func validateKey(key string) error {
	if key == "" {
		return errors.New("secrets: key is empty")
	}

	if strings.ContainsAny(key, "\x00\n\r") {
		return fmt.Errorf("secrets: key %q contains a control character", key)
	}
	return nil
}

func Get(key string) (string, error) {
	if err := validateKey(key); err != nil {
		return "", err
	}
	value, err := keyring.Get(service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return value, nil
}

func Set(key, value string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := keyring.Set(service, key, value); err != nil {
		return err
	}
	return nil
}

func Delete(key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := keyring.Delete(service, key); err != nil {
		return err
	}
	return nil

}
