package cache

import (
	"github.com/valkey-io/valkey-go"
)

func NewValkey(addr string, username string, password string) (valkey.Client, error) {
	return valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
		Username: username,
		Password: password,
	})
}