package service

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestCheckSigFailedErr_Error(t *testing.T) {
	type fields struct {
		Msg string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSigFailedErr{
				Msg: tt.fields.Msg,
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("CheckSigFailedErr.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_Sign(t *testing.T) {
	msg := "hello world"
	sig := CertificateAuthorityX509.Sign([]byte(msg))
	fmt.Println(len(sig), sig)
}

func TestCAX509_VerifySignature(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		sig []byte
		msg []byte
		Id  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.VerifySignature(tt.args.sig, tt.args.msg, tt.args.Id); got != tt.want {
				t.Errorf("CAX509.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_Encode(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		msg []byte
		Id  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := c.Encode(tt.args.msg, tt.args.Id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAX509.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_Decode(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		EncodeMsg []byte
		Id        string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := c.Decode(tt.args.EncodeMsg, tt.args.Id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAX509.Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_GetCerts(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]x509.Certificate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.GetCerts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAX509.GetCerts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_DelCert(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		Id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if err := ca.DelCert(tt.args.Id); (err != nil) != tt.wantErr {
				t.Errorf("CAX509.DelCert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCAX509_GetSeeds(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.GetSeeds(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAX509.GetSeeds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_VerifyCertificate(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.VerifyCertificate(tt.args.data); got != tt.want {
				t.Errorf("CAX509.VerifyCertificate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_GetLocalCertificate(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	tests := []struct {
		name   string
		fields fields
		want   *x509.Certificate
		want1  []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			got, got1 := ca.GetLocalCertificate()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CAX509.GetLocalCertificate() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CAX509.GetLocalCertificate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCAX509_GetNetworkSize(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.GetNetworkSize(); got != tt.want {
				t.Errorf("CAX509.GetNetworkSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_GetF(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.GetF(); got != tt.want {
				t.Errorf("CAX509.GetF() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_Exits(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.Exits(tt.args.id); got != tt.want {
				t.Errorf("CAX509.Exits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_Check(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.Check(tt.args.n); got != tt.want {
				t.Errorf("CAX509.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCAX509_IsLeaderNode(t *testing.T) {
	type fields struct {
		Mutex             sync.Mutex
		Certificates      map[string]x509.Certificate
		CertificatesOrder []*x509.Certificate
		NodeId            int64
		PrivateKey        *rsa.PrivateKey
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CAX509{
				Mutex:             tt.fields.Mutex,
				Certificates:      tt.fields.Certificates,
				CertificatesOrder: tt.fields.CertificatesOrder,
				NodeId:            tt.fields.NodeId,
				PrivateKey:        tt.fields.PrivateKey,
			}
			if got := ca.IsLeaderNode(tt.args.id); got != tt.want {
				t.Errorf("CAX509.IsLeaderNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadPrivateKey2(t *testing.T) {
	tests := []struct {
		name string
		want *rsa.PrivateKey
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadPrivateKey2(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadPrivateKey2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadCertificate2(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want *x509.Certificate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadCertificate2(tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadCertificate2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loadCertificate2Bytes(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loadCertificate2Bytes(tt.args.fileName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadCertificate2Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDigest2(t *testing.T) {
	type args struct {
		msg []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDigest2(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDigest2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDigest2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_insertCertificateByOrder(t *testing.T) {
	type args struct {
		certs []*x509.Certificate
		cert  *x509.Certificate
	}
	tests := []struct {
		name string
		args args
		want []*x509.Certificate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := insertCertificateByOrder(tt.args.certs, tt.args.cert); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("insertCertificateByOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
