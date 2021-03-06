// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcutil_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/monetas/btcutil"
	//"github.com/davecgh/go-spew/spew"
)

// TestNewTLSCertPair ensures the NewTLSCertPair function works as expected.
func TestNewTLSCertPair(t *testing.T) {
	// Certs don't support sub-second precision, so truncate it now to
	// ensure the checks later don't fail due to nanosecond precision
	// differences.
	validUntil := time.Unix(time.Now().Add(10*365*24*time.Hour).Unix(), 0)
	org := "test autogenerated cert"
	extraHosts := []string{"testtlscert.bogus", "127.0.0.1"}
	cert, key, err := btcutil.NewTLSCertPair(org, validUntil, extraHosts)
	if err != nil {
		t.Fatalf("failed with unexpected error: %v", err)
	}

	// Ensure the PEM-encoded cert that is returned can be decoded.
	pemCert, _ := pem.Decode(cert)
	if pemCert == nil {
		t.Fatalf("pem.Decode was unable to decode the certificate")
	}

	// Ensure the PEM-encoded key that is returned can be decoded.
	pemKey, _ := pem.Decode(key)
	if pemCert == nil {
		t.Fatalf("pem.Decode was unable to decode the key")
	}

	// Ensure the DER-encoded key bytes can be successfully parsed.
	_, err = x509.ParseECPrivateKey(pemKey.Bytes)
	if err != nil {
		t.Fatalf("failed with unexpected error: %v", err)
	}

	// Ensure the DER-encoded cert bytes can be successfully into an X.509
	// certificate.
	x509Cert, err := x509.ParseCertificate(pemCert.Bytes)
	if err != nil {
		t.Fatalf("failed with unexpected error: %v", err)
	}

	// Ensure the specified organization is correct.
	x509Orgs := x509Cert.Subject.Organization
	if len(x509Orgs) == 0 || x509Orgs[0] != org {
		x509Org := "<no organization>"
		if len(x509Orgs) > 0 {
			x509Org = x509Orgs[0]
		}
		t.Fatalf("generated cert organization field mismatch, got "+
			"'%v', want '%v'", x509Org, org)
	}

	// Ensure the specified valid until value is correct.
	if !x509Cert.NotAfter.Equal(validUntil) {
		t.Fatalf("generated cert valid until field mismatch, got %v, "+
			"want %v", x509Cert.NotAfter, validUntil)
	}

	// Ensure the specified extra hosts are present.
	for _, host := range extraHosts {
		if err := x509Cert.VerifyHostname(host); err != nil {
			t.Fatalf("failed to verify extra host '%s'", host)
		}
	}

	// Ensure the cert can be use for the intended purposes.
	if !x509Cert.IsCA {
		t.Fatal("generated cert is not a certificate authority")
	}
	if x509Cert.KeyUsage&x509.KeyUsageKeyEncipherment == 0 {
		t.Fatal("generated cert can't be used for key encipherment")
	}
	if x509Cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		t.Fatal("generated cert can't be used for digital signatures")
	}
	if x509Cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		t.Fatal("generated cert can't be used for signing other certs")
	}
	if !x509Cert.BasicConstraintsValid {
		t.Fatal("generated cert does not have valid basic constraints")
	}
}
