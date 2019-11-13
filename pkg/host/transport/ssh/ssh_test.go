package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	privateKey, err := generateRSAPrivateKey()
	if err != nil {
		t.Fatalf("generating RSA key failed")
	}

	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
		PrivateKey:        privateKey,
	}
	if _, err := c.New(); err != nil {
		t.Fatalf("creating new SSH object should succeed, got: %s", err)
	}
}

func TestNewValidate(t *testing.T) {
	c := &Config{}
	if _, err := c.New(); err == nil {
		t.Fatalf("creating new SSH object should validate it")
	}
}

func TestValidateRequireAddress(t *testing.T) {
	c := &Config{
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require address field")
	}
}

func TestValidateRequireUser(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		Password:          "foo",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require user field")
	}
}

func TestValidateRequireAuthMethod(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require at least one authentication method")
	}
}

func TestValidateRequireConnectionTimeout(t *testing.T) {
	c := &Config{
		Address:  "localhost",
		User:     "root",
		Password: "foo",
		Port:     22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require connection timeout field")
	}
}

func TestValidateRequirePort(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30s",
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should require port field")
	}
}

func TestValidateParseConnectionTimeout(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		Password:          "foo",
		ConnectionTimeout: "30",
		Port:              22,
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should parse connection timeout")
	}
}

func TestValidateParsePrivateKey(t *testing.T) {
	c := &Config{
		Address:           "localhost",
		User:              "root",
		ConnectionTimeout: "30s",
		Port:              22,
		PrivateKey:        "foo",
	}
	if _, err := c.New(); err == nil {
		t.Fatalf("validating SSH configuration should parse private key")
	}
}

func generateRSAPrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("generating key failed: %w", err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return string(pem.EncodeToMemory(&privBlock)), nil
}
