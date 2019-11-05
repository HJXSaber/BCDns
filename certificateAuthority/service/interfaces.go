package service

import (
	"BCDns_0.1/certificateAuthority/model"
	"BCDns_0.1/protos"
)

//CAInterface
type CAInterface interface {
	LoadCertificates()
	RefreshCertificates()
	GenerateCertificate() model.Certificate
	Revoke(certificate model.Certificate) bool
	Propagate(certificate model.Certificate)
	getNetworkSiz() int64
	Encrypt(message protos.Transaction) ([]byte, error)
	Decrypt(message []byte) (protos.Transaction, error)
	SignatureCrt(certificate model.Certificate) ([]byte, error)
	SignatureTx(transaction protos.Transaction) ([]byte, error)
	ValidateSignature(certificate model.Certificate) (bool, error)
}

type CertificateInterface interface {
	ValidateSignature([]byte, []byte) (bool, error)
	GetName() (string, error)
	GetAddr() string
}
