package auth

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

type Keyring struct {
	service string
}

func NewKeyring() *Keyring {
	return &Keyring{service: "megahorn"}
}

func (k *Keyring) ServiceName() string {
	return k.service
}

func (k *Keyring) Key(platform, name string) string {
	return fmt.Sprintf("%s:%s:%s", k.service, platform, name)
}

func (k *Keyring) Set(platform, name, value string) error {
	return keyring.Set(k.service, k.Key(platform, name), value)
}

func (k *Keyring) Get(platform, name string) (string, error) {
	return keyring.Get(k.service, k.Key(platform, name))
}

func (k *Keyring) Delete(platform, name string) error {
	return keyring.Delete(k.service, k.Key(platform, name))
}
