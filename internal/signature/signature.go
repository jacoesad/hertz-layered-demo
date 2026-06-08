package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Config struct {
	Enabled bool
	Secret  string
	Header  string
}

type Verifier interface {
	Enabled() bool
	Header() string
	Sign(method string, path string) string
	Verify(method string, path string, signature string) error
}

type verifier struct {
	enabled bool
	secret  string
	header  string
}

func NewVerifier(cfg Config) (Verifier, error) {
	header := cfg.Header
	if header == "" {
		header = "X-Signature"
	}
	if cfg.Enabled && cfg.Secret == "" {
		return nil, fmt.Errorf("signature secret is required when signature is enabled")
	}
	return &verifier{
		enabled: cfg.Enabled,
		secret:  cfg.Secret,
		header:  header,
	}, nil
}

func (v *verifier) Enabled() bool {
	return v != nil && v.enabled
}

func (v *verifier) Header() string {
	if v == nil || v.header == "" {
		return "X-Signature"
	}
	return v.header
}

func (v *verifier) Sign(method string, path string) string {
	mac := hmac.New(sha256.New, []byte(v.secret))
	_, _ = mac.Write([]byte(method))
	_, _ = mac.Write([]byte("\n"))
	_, _ = mac.Write([]byte(path))
	return hex.EncodeToString(mac.Sum(nil))
}

func (v *verifier) Verify(method string, path string, signature string) error {
	if !v.Enabled() {
		return nil
	}
	if signature == "" {
		return fmt.Errorf("signature is required")
	}
	expected := v.Sign(method, path)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("signature is invalid")
	}
	return nil
}
