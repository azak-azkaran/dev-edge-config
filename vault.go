package main

import (
	"errors"

	vault "github.com/hashicorp/vault/api"
)

func GetSecret(config *vault.Config, token string, path string) (map[string]interface{}, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	logical := client.Logical()
	secret, err := logical.Read(path)
	if err != nil {
		return nil, err
	}

if _, ok := secret.Data["data"]; ok {
		data := secret.Data["data"].(map[string]interface{})
		if len(data) == 0 {
			return nil, errors.New("Data of secret with path: " + path + " is empty")
		}
		return data, nil
	} else {
		return secret.Data, nil
	}
}
func SealStatus(config *vault.Config) (*vault.SealStatusResponse, error) {
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	sys := client.Sys()
	respones, err := sys.SealStatus()
	if err != nil {
		return nil, err
	}
	return respones, nil
}
