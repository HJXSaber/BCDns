package service

import (
	"BCDns_0.1/certificateAuthority/model"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"reflect"
	"testing"
	"time"
)

//func TestCAImpl_LoadCertificates(t *testing.T) {
//	os.Setenv("BCDNS_CERTIFICATE_PATH", "./")
//	ca := CAImpl{
//		Certificates: map[string]*model.Certificate{},
//	}
//	ca.LoadCertificates()
//	for _, cert := range ca.Certificates {
//		fmt.Println(*cert)
//		res, err := utils.Convert(*cert, model.PbCertT, true)
//		if err != nil {
//			log.Println(err)
//			return
//		}
//		certProto := res.(protos.Certificate)
//		fmt.Println(certProto)
//		res, err = utils.Convert(certProto, model.ModelCertT, true)
//		if err != nil {
//			log.Println(err)
//			return
//		}
//		certRecover := res.(model.Certificate)
//		fmt.Println(certRecover)
//	}
//}

type TestType struct {
	T1 *T1
}

type T1 struct {
	A int
}

type HandrI func(i int)

var h HandrI = func(i int) {
	fmt.Println(i)
}

var (
	handlerInt = reflect.TypeOf(*new(HandrI))
)

func TestCAImpl_LoadCertificates2(t *testing.T) {
	path := "cert.crt"
	caFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	temp := make([]byte, len(caFile) * 4 / 3)
	base64.StdEncoding.Decode(temp, caFile)

	//decode the content of cert by base64
	caBlock, _ := pem.Decode(caFile)
	cert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(cert)
}

//func TestCAImpl_GenerateCertificate(t *testing.T) {
//	ca := CAImpl{}
//	ca.GenerateCertificate()
//}

func TestWhatever(t *testing.T) {
	var subject model.Subject
	fmt.Println(reflect.TypeOf(subject))
}

func TestConvert(t *testing.T) {
	//protoCert := model.Certificate{
	//	RawCertificate: model.RawCertificate{
	//		Version:1,
	//		SerialNumber: big.NewInt(1),
	//		PublicKey: model.PublicKey{
	//			N:big.NewInt(12),
	//		},
	//	},
	//}
	//cert, err := utils.Convert(protoCert, reflect.TypeOf(protos.Certificate{}), true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//c := cert.(protos.Certificate)
	//fmt.Println(c)
	//
	//certBack, err := utils.Convert(c, reflect.TypeOf(model.Certificate{}), true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//cc := certBack.(model.Certificate)
	//fmt.Println(cc)
	/*{
		i := 1
		time.AfterFunc(2 * time.Second, func() {
			fmt.Println(i)
		})
	}
	fmt.Println(100)
	time.Sleep(3 * time.Second)*/

	max := new(big.Int).Lsh(big.NewInt(1),128)  //把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max)   //返回在 [0, max) 区间均匀随机分布的一个随机值
	subject := pkix.Name{   //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
		Organization:       []string{"Manning Publications Co."},
		OrganizationalUnit: []string{"Books"},
		CommonName:         "Go Web Programming",
	}
	template := x509.Certificate{
		SerialNumber:   serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject:        subject,
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(365 * 24 *time.Hour),
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, //KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, // 密钥扩展用途的序列
		IPAddresses:    []net.IP{net.ParseIP("127.0.0.1")},
	}
	pk, _ := rsa.GenerateKey(rand.Reader, 2048) //生成一对具有指定字位数的RSA密钥
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk) //DER 格式
	cert, _ := x509.ParseCertificate(derBytes)
	block := pem.Block{
		Type: "certificate",
		Bytes: derBytes,
	}
	file, _ := os.Create("./test.pem")
	_ = pem.Encode(file, &block)
	file.Close()
	fileInfo, _ := os.Stat("./test.pem")
	file, _ = os.Open(fileInfo.Name())
	content := make([]byte, fileInfo.Size())
	_, _ = file.Read(content)
	block2, _ := pem.Decode(content)
	certRecover , _ := x509.ParseCertificate(block2.Bytes)
	fmt.Println(certRecover.Equal(cert))

	data, _ := json.Marshal(cert)
	fmt.Println(data)
	_ = json.Unmarshal(data, certRecover)
	fmt.Println(cert.Equal(certRecover))
}

func TestGenerateCertificate(t *testing.T) {
	rootCert := loadCertificate2("../conf/s1/LocalCertificate.crt")
	fmt.Println(rootCert)
}

func TestCAX509_VerifyCertificate(t *testing.T) {
	msg := []byte("I am zzy")
	sig := CertificateAuthorityX509.Sign(msg)
	fmt.Println(len(sig), sig)
}