package flexkube

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/flexkube/libflexkube/pkg/pki"
	"github.com/flexkube/libflexkube/pkg/types"
)

func certificateMarshal(c *pki.Certificate) interface{} {
	return map[string]interface{}{
		"organization":      c.Organization,
		"rsa_bits":          c.RSABits,
		"validity_duration": c.ValidityDuration,
		"renew_threshold":   c.RenewThreshold,
		"common_name":       c.CommonName,
		"ca":                c.CA,
		"key_usage":         stringSliceToInterfaceSlice(c.KeyUsage),
		"ip_addresses":      stringSliceToInterfaceSlice(c.IPAddresses),
		"dns_names":         stringSliceToInterfaceSlice(c.DNSNames),
		"x509_certificate":  c.X509Certificate,
		"public_key":        c.PublicKey,
		"private_key":       c.PrivateKey,
	}
}

func certificateUnmarshal(i interface{}) *pki.Certificate {
	c := &pki.Certificate{}

	if i == nil {
		return c
	}

	j, ok := i.([]interface{})

	if !ok || len(j) != 1 {
		return c
	}

	k := j[0].(map[string]interface{})

	if len(j) == 0 {
		return c
	}

	return &pki.Certificate{
		Organization:     k["organization"].(string),
		RSABits:          k["rsa_bits"].(int),
		ValidityDuration: k["validity_duration"].(string),
		RenewThreshold:   k["renew_threshold"].(string),
		CommonName:       k["common_name"].(string),
		CA:               k["ca"].(bool),
		KeyUsage:         stringListUnmarshal(k["key_usage"]),
		IPAddresses:      stringListUnmarshal(k["ip_addresses"]),
		DNSNames:         stringListUnmarshal(k["dns_names"]),
		X509Certificate:  types.Certificate(k["x509_certificate"].(string)),
		PublicKey:        k["public_key"].(string),
		PrivateKey:       types.PrivateKey(k["private_key"].(string)),
	}
}

func certificateBlockSchema(computed bool) *schema.Schema {
	return optionalBlock(computed, certificateSchema)
}

func certificateSchema(computed bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization":      optionalString(computed),
		"rsa_bits":          optionalInt(computed),
		"validity_duration": optionalString(computed),
		"renew_threshold":   optionalString(computed),
		"common_name":       optionalString(computed),
		"ca":                optionalBool(computed),
		"key_usage":         optionalStringList(computed),
		"ip_addresses":      optionalStringList(computed),
		"dns_names":         optionalStringList(computed),
		"x509_certificate":  optionalString(computed),
		"public_key":        optionalString(computed),
		"private_key":       sensitiveString(computed),
	}
}
