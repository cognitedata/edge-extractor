package internal

import "os"

type SecretManager struct {
	Key     string
	Secrets map[string]string // map of decrypted secrets
}

func NewSecretManager(key string) *SecretManager {
	return &SecretManager{Key: key, Secrets: map[string]string{}}
}

func (sm *SecretManager) LoadEncryptedSecrets(secrets map[string]string) error {
	var err error
	for k, v := range secrets {
		sm.Secrets[k], err = DecryptString(sm.Key, v)
	}
	return err
}

// returns secret either from internal secret store or from ENV variable if it is not found in the store.
// If secret is not found in ENV variable, returns key (plain text)
func (sm *SecretManager) GetSecret(key string) string {
	secret, ok := sm.Secrets[key]
	if !ok {
		secret = os.Getenv(key)
		if secret == "" {
			return key
		}
		return secret
	}
	return secret
}

func (sm *SecretManager) GetEncryptedSecrets() map[string]string {
	encryptedSecrets := map[string]string{}
	for k, v := range sm.Secrets {
		encryptedSecrets[k], _ = EncryptString(sm.Key, v)
	}
	return encryptedSecrets
}
