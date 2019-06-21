package utils

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"reflect"
)

var (
	x509Certificate = reflect.TypeOf(x509.Certificate{})
	x509CertificatePtr = reflect.TypeOf(&x509.Certificate{})

)

type GetIDError struct {
	Msg string
}

func (err GetIDError) Error() string {
	return err.Msg
}

func GetCertId(cert interface{}) (string, error) {
	switch reflect.TypeOf(cert) {
	case x509CertificatePtr:
		certificate := cert.(*x509.Certificate)
		en, err := MakeBigInt(certificate.SerialNumber)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		id := make([]byte, en.Len())
		en.Encode(id)
		return base64.StdEncoding.EncodeToString(id), nil
	case x509Certificate:
		certificate := cert.(x509.Certificate)
		en, err := MakeBigInt(certificate.SerialNumber)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		id := make([]byte, en.Len())
		en.Encode(id)
		return base64.StdEncoding.EncodeToString(id), nil
	default:
		fmt.Println("Type is ", reflect.TypeOf(cert))
		return "", GetIDError{"Unknown certificate type"}
	}
}