package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// Hash identifies the effective configuration without credentials.
func (c *Config) Hash() (string, error) {
	if c == nil {
		return "", &FieldError{
			Field:  "config",
			Reason: "must not be nil",
		}
	}

	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("config: encode for hash: %w", err)
	}

	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum[:6]), nil
}
