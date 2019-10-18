package service

import (
	"crypto/x509"
)

type CAX509Interface interface {
	Sign(msg []byte) []byte
	VerifySignature(sig, msg []byte, Id string) bool
	Encode(msg []byte, Id string) []byte
	Decode(EncodeMsg []byte, Id string) []byte
	GetCerts() map[string]x509.Certificate
	AddCert(data []byte) error
	DelCert(Id string) error
	GetSeeds() []string
	VerifyCertificate(data []byte) bool
	GetLocalCertificate() (*x509.Certificate, []byte)
	GetNetworkSize() int
	GetF() int
}

