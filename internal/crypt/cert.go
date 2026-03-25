package crypt

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type Cert struct {
	PemPath   string `binding:"required"`
	KeyPath   string `binding:"required"`
	PemBytes  []byte `binding:"required"`
	KeyBytes  []byte `binding:"required"`
	CertBytes []byte `binding:"required"`
}

func GenerateKeyECDSA() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func GenerateKeyRSA(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	//fmt.Println("Private Key generated")
	return privateKey, nil
}

// ----------------------------------------------------------------------------------------------------------------------------------------------

func CreateCertificateECDSA(key *ecdsa.PrivateKey) (*[]byte, error) {
	template := getTemplate()
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, getKeyType(key), key)
	if err != nil {
		return nil, err
	}
	return &certBytes, nil
}

func CreateCertificateRSA(key *rsa.PrivateKey) (*[]byte, error) {
	template := getTemplate()
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, getKeyType(key), key)
	if err != nil {
		return nil, err
	}
	return &certBytes, nil
}

// ----------------------------------------------------------------------------------------------------------------------------------------------

func getTemplate() *x509.Certificate {

	//sn := big.NewInt(1)
	sn := big.NewInt(time.Now().Unix())
	//fmt.Println("sn=", sn)

	template := x509.Certificate{
		SerialNumber: sn,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365 * 100),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	return &template
}

func getKeyType(key interface{}) interface{} {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func encodePem(pemBuf *bytes.Buffer, certBytes []byte) error {
	err := pem.Encode(pemBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return err
	}
	return nil
}

func encodeKey(pemBuf *bytes.Buffer, key interface{}) error {
	pemBlock, err := readPemForKey(key)
	if err != nil {
		return err
	}
	err = pem.Encode(pemBuf, pemBlock)
	if err != nil {
		return err
	}
	return nil
}

func readPemForKey(key interface{}) (*pem.Block, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal ECDSA private key: %w", err)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, errors.New("failed to determine key type")
	}
}

func createKeyFile(certPath, certName, certSfx string, b []byte) (string, error) {

	// certSfx = ".key"

	path := path.Join(certPath, certName+certSfx)

	err := os.WriteFile(path, b, 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

// ------------------------------------------------------------------------------------------------------------------------

func MakeRSAKeyPairAndSave(certPath, certName string) (*Cert, error) {

	if len(certPath) == 0 {
		certPath = ""
	}
	if len(certName) == 0 {
		certName = "server"
	}

	key, err := GenerateKeyRSA(2048)
	if err != nil {
		return nil, err
	}

	certBytes, err := CreateCertificateRSA(key)
	pemBuf := &bytes.Buffer{}
	err = encodePem(pemBuf, *certBytes)
	if err != nil {
		return nil, err
	}

	var cert Cert
	cert.CertBytes = *certBytes

	certPem := pemBuf
	pemPath, err := createKeyFile(certPath, certName, ".pem", certPem.Bytes())
	if err != nil {
		return nil, err
	}

	cert.PemBytes = pemBuf.Bytes()
	cert.PemPath = pemPath

	pemBuf.Reset()
	err = encodeKey(pemBuf, key)
	if err != nil {
		return nil, err
	}
	cert.PemBytes = pemBuf.Bytes()

	certKey := pemBuf
	keyPath, err := createKeyFile(certPath, certName, ".key", certKey.Bytes())
	if err != nil {
		return nil, err
	}
	cert.KeyPath = keyPath

	fmt.Println("")

	return &cert, nil

}
