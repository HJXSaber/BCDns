package model

import (
	"BCDns_0.1/protos"
	"crypto"
	"crypto/rsa"
	"encoding/asn1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"log"
	"math/big"
	"reflect"
	"strconv"
)

var (
	ModelCertT = reflect.TypeOf(Certificate{})
	PbCertT = reflect.TypeOf(protos.Certificate{})
)

//certificate begin
type Certificate struct {
	RawCertificate
	Signatures []Signature
}

func (cert Certificate) ValidateSignature(content, signature []byte) (bool, error) {
	err := rsa.VerifyPKCS1v15(&rsa.PublicKey{
		N: cert.N,
		E: int(cert.E),
	}, crypto.SHA256, content, signature)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

func (cert Certificate) GetName() (string, error) {
	return cert.Subject.String()
}

func (cert Certificate) GetAddr() string {
	return cert.Addr.String()
}

func (cert Certificate) MarshalBinary() (data []byte, err error) {
	content, err := asn1.Marshal(cert)
	if err != nil {
		return []byte{}, nil
	}
	return content, nil
}

type RawCertificate struct {
	Version int64
	SerialNumber *big.Int
	ValidFrom, ValidTo timestamp.Timestamp
	Subject
	PublicKey
	Addr
	NetworkSize int64
}

func (certRaw RawCertificate) MarshalBinary() (data []byte, err error) {
	content, err := asn1.Marshal(certRaw)
	if err != nil {
		return []byte{}, nil
	}
	return content, nil
}

type Subject struct {
	Country, Organization, OrganizationalUnit []string
	Locality, Province                        []string
	StreetAddress, PostalCode                 []string
	SerialNumber, CommonName                  string
}

func (subject Subject) String() (string, error) {
	hash := crypto.MD5.New()
	subjectByte, err := asn1.Marshal(subject)
	if err != nil {
		//TODO
		return "", err
	}
	hash.Write(subjectByte)
	digest := hash.Sum(nil)
	return string(digest), nil
}

type PublicKey struct {
	N	*big.Int
	E	int64
}

type Signature struct {
	Subject
	Signature            []byte
}

type Addr struct {
	Ip string
	Port int32
}

func (addr Addr) String() string {
	return addr.Ip + ":" + strconv.FormatInt(int64(addr.Port), 10)
}

//certificate end

//Replaced by utils.Convert method
/*func CertToProtoCert(certificate Certificate) *protos.Certificate {
	signatures := make(map[string]*protos.Signature)
	for name, sig := range certificate.Signatures {
		signatures[name] = &protos.Signature{
			Subject: &protos.Subject{
				Country:            sig.Country,
				Organization:       sig.Organization,
				OrganizationalUnit: sig.OrganizationalUnit,
				Locality:           sig.Locality,
				Province:           sig.Province,
				StreetAddress:      sig.StreetAddress,
				PostalCode:         sig.PostalCode,
				SerialNumber:       sig.SerialNumber,
				CommonName:         sig.CommonName,
			},
			Signature: sig.Signature,
		}
	}
	serialNumberEn, err := utils.MakeBigInt(certificate.SerialNumber)
	if err != nil {
		//TPDP
		log.Fatal(err)
	}
	serialNumber := make([]byte, serialNumberEn.Len())
	serialNumberEn.Encode(serialNumber)
	nEn, err := utils.MakeBigInt(certificate.N)
	if err != nil {
		//TPDP
		log.Fatal(err)
	}
	pN := make([]byte, nEn.Len())
	nEn.Encode(pN)
	return &protos.Certificate{
		RawCertificate: &protos.RawCertificate{
			Version:      certificate.Version,
			SerialNumber: serialNumber,
			ValidFrom:    &certificate.ValidFrom,
			ValidTo:      &certificate.ValidTo,
			Subject: &protos.Subject{
				Country:            certificate.Country,
				Organization:       certificate.Organization,
				OrganizationalUnit: certificate.OrganizationalUnit,
				Locality:           certificate.Locality,
				Province:           certificate.Province,
				StreetAddress:      certificate.StreetAddress,
				PostalCode:         certificate.PostalCode,
				SerialNumber:       certificate.RawCertificate.Subject.SerialNumber,
				CommonName:         certificate.CommonName,
			},
			PublicKey: &protos.PublicKey{
				N: pN,
				E: certificate.E,
			},
			Addr: &protos.Addr{
				Ip: certificate.Ip,
				Port: certificate.Port,
			},
			NetworkSize: certificate.NetworkSize,
		},
		Signatures: signatures,
	}
}

func ProtoCertToCert(cert protos.Certificate) (*Certificate, error) {
	signatures := make(map[string]Signature)
	for name, sig := range cert.Signatures {
		signatures[name] = Signature{
			Subject: Subject{
				Country:            sig.Subject.Country,
				Organization:       sig.Subject.Organization,
				OrganizationalUnit: sig.Subject.OrganizationalUnit,
				Locality:           sig.Subject.Locality,
				Province:           sig.Subject.Province,
				StreetAddress:      sig.Subject.StreetAddress,
				PostalCode:         sig.Subject.PostalCode,
				SerialNumber:       sig.Subject.SerialNumber,
				CommonName:         sig.Subject.CommonName,
			},
			Signature: sig.Signature,
		}
	}
	serialNumber, err := utils.ParseBigInt(cert.RawCertificate.SerialNumber)
	if err != nil {
		//TODO
		return nil, err
	}
	pN, err := utils.ParseBigInt(cert.RawCertificate.PublicKey.N)
	if err != nil {
		//TODO
		return nil, err
	}
	return &Certificate{
		RawCertificate: RawCertificate{
			Version: cert.RawCertificate.Version,
			SerialNumber: serialNumber,
			ValidFrom: *cert.RawCertificate.ValidFrom,
			ValidTo: *cert.RawCertificate.ValidTo,
			Subject: Subject{
				Country: cert.RawCertificate.Subject.Country,
				Organization: cert.RawCertificate.Subject.Organization,
				OrganizationalUnit: cert.RawCertificate.Subject.OrganizationalUnit,
				Locality: cert.RawCertificate.Subject.Locality,
				Province: cert.RawCertificate.Subject.Province,
				StreetAddress: cert.RawCertificate.Subject.StreetAddress,
				PostalCode: cert.RawCertificate.Subject.PostalCode,
				SerialNumber: cert.RawCertificate.Subject.SerialNumber,
				CommonName: cert.RawCertificate.Subject.CommonName,
			},
			PublicKey: PublicKey{
				N: pN,
				E: cert.RawCertificate.PublicKey.E,
			},
			Addr: Addr{
				Ip: cert.RawCertificate.Addr.Ip,
				Port: cert.RawCertificate.Addr.Port,
			},
			NetworkSize: cert.RawCertificate.NetworkSize,
		},
		Signatures: signatures,
	}, nil
}*/