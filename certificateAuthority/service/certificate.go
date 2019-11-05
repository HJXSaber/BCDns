package service

//
//import (
//	"BCDns_0.1/certificateAuthority/model"
//	model2 "BCDns_0.1/network/model"
//	pb "BCDns_0.1/protos"
//	"BCDns_0.1/utils"
//	"context"
//	"crypto"
//	"crypto/rand"
//	"crypto/rsa"
//	"encoding/asn1"
//	"encoding/pem"
//	"github.com/golang/protobuf/ptypes"
//	"google.golang.org/grpc"
//	"io/ioutil"
//	"log"
//	"math/big"
//	"os"
//	"reflect"
//	"regexp"
//	"time"
//)
//
//var (
//	RootCertPath string = "./RootCert.pem.cert"
//	RootPrivateKey string = "./RootOri.pem.key"
//	CertificateAuthority CAImpl
//)
//
//type ReviewItem struct {
//	Cert *model.Certificate
//	Responses []*model.Certificate
//	Timer *time.Timer
//}
//
////CA implementation
//type CAImpl struct {
//	Certificates map[string]*model.Certificate
//	certToAudit map[string]*model.Certificate
//	CertWaitReview *ReviewItem
//	CertUpdate map[string]*model.Certificate
//}
//
//func (ca *CAImpl) LoadCertificates() {
//	//files, err := ioutil.ReadDir(bcDns.Dns.Config.CAPath)
//	files, err := ioutil.ReadDir(os.Getenv("BCDNS_CERTIFICATE_PATH"))
//	if err != nil {
//		//TODO
//	}
//	for _, file := range files {
//		if file.IsDir() {
//			continue
//		} else if ok, _ := regexp.MatchString(`.*\.cert$`, file.Name()); ok {
//			if _, exit := ca.Certificates[file.Name()]; !exit {
//				ca.Certificates[file.Name()], err = loadCertificate(file.Name())
//			}
//		}
//	}
//}
//
//func (*CAImpl) RefreshCertificates() {
//	panic("implement me")
//}
//
//func (*CAImpl) GenerateCertificate() model.Certificate {
//	var parent *model.Certificate = nil
//	if utils.Exists(RootCertPath) {
//		parent, _  = loadCertificate(RootCertPath)
//	}
//	return *createCert(parent)
//}
//
//func (*CAImpl) Revoke(certificate model.Certificate) bool {
//	panic("implement me")
//}
//
//func (ca *CAImpl) Propagate(certificate model.Certificate) {
//	if ca.CertWaitReview != nil {
//		if certificate.SerialNumber.Cmp(ca.CertWaitReview.Cert.SerialNumber) != 1 {
//			//TODO
//			log.Println("Propagate: Cert is invalid")
//			return
//		}
//		ca.CertWaitReview.Timer.Stop()
//		//ca.CertWaitReview = nil
//	}
//	ca.CertWaitReview = &ReviewItem{
//		Cert: &certificate,
//		Responses: []*model.Certificate{},
//		Timer: time.AfterFunc(24 * time.Hour, func() {
//			if len(ca.CertWaitReview.Responses) == len(ca.Certificates) {
//				//Audit successfully
//			} else {
//				ca.CertWaitReview = nil
//			}
//		}),
//	}
//	for _, node := range model2.P2PNet.Network.Members() {
//		go func() {
//			var options []grpc.DialOption
//			//TODO communicate with TLS
//			conn, err := grpc.Dial(node.Address(), options...)
//			if err != nil {
//				//TODO
//				log.Fatal("Propagate:", err)
//			}
//			defer conn.Close()
//			client := pb.NewCertificateAuthorityClient(conn)
//			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//			defer cancel()
//			res, err := utils.Convert(certificate, model.PbCertT, true)
//			if err != nil {
//				log.Fatal("Propagate:", err)
//			}
//			cert := res.(pb.Certificate)
//			_, err = client.ProposalCert(ctx, &cert)
//			if err != nil {
//				log.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
//			}
//		}()
//	}
//}
//
//func (ca *CAImpl) getNetworkSiz() (res int64) {
//	res = -1
//	cert, _ := loadCertificate(RootCertPath)
//	if cert != nil {
//		res = cert.NetworkSize
//	} else if len(ca.Certificates) != 0 {
//		for _, cert = range ca.Certificates {
//			res = cert.NetworkSize
//			break
//		}
//	}
//	return
//}
//
//func (ca *CAImpl) Encrypt(message pb.Transaction) ([]byte, error) {
//	panic("implement me")
//}
//
//func (ca *CAImpl) Decrypt(message []byte) (pb.Transaction, error) {
//	panic("implement me")
//}
//
//func (ca *CAImpl) SignatureCrt(certificate model.Certificate) ([]byte, error) {
//	digest, err := certificate.RawCertificate.MarshalBinary()
//	if err != nil {
//		return []byte{}, err
//	}
//	privateK, err := loadPrivateKey(RootPrivateKey)
//	if err != nil {
//		return []byte{}, err
//	}
//	signature, err := privateK.Sign(rand.Reader, digest, crypto.SHA256)
//	if err != nil {
//		return []byte{}, err
//	}
//	return signature, nil
//}
//
//func (ca *CAImpl) SignatureTx(transaction pb.Transaction) ([]byte, error) {
//	panic("implement me")
//}
//
//type SigValidateError struct {
//	Msg string
//}
//
//func (this SigValidateError) Error() string {
//	return this.Msg
//}
//
//func (ca *CAImpl) ValidateSignature(certificate model.Certificate) (bool, error)  {
//	if int64(len(certificate.Signatures)) != certificate.NetworkSize {
//		return false, SigValidateError{"Insufficient signature"}
//	}
//	for _, signature := range certificate.Signatures {
//		subName, err := signature.Subject.String()
//		if err != nil {
//			//TODO
//			log.Println("Validate signatures failed")
//			return false, err
//		}
//		if cert, ok := ca.Certificates[subName]; ok {
//			content, err := certificate.RawCertificate.MarshalBinary()
//			if err != nil {
//				log.Println(err)
//				return false, err
//			}
//			res, err := cert.ValidateSignature(content, signature.Signature)
//			if err != nil {
//				log.Println(err)
//				return false, err
//			}
//			if !res {
//				log.Println("Incorrect signature")
//				return false, SigValidateError{"Incorrect signature"}
//			}
//		}
//	}
//	return true, nil
//}
//
//// Pem Block
//// x509.CreateCertificate()
//
//// The encoded form is:
////    -----BEGIN Type-----
////    Headers
////    base64-encoded Bytes
////    -----END Type-----
//func createCert(parent *model.Certificate) *model.Certificate {
//	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//	stream, _ := asn1.Marshal(*privateKey)
//	block := &pem.Block{
//		Type:"RSA private key",
//		Bytes:stream,
//	}
//	file, err := os.Create(RootPrivateKey)
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//	err = pem.Encode(file, block)
//	var serialNumber *big.Int
//	if parent != nil {
//		serialNumber = parent.RawCertificate.SerialNumber
//	} else {
//		serialNumber = big.NewInt(0)
//	}
//	validFrom, err := ptypes.TimestampProto(time.Now())
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//	validTo, err := ptypes.TimestampProto(time.Now().Add(365 * 24 * time.Hour))
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//	subject := model.Subject{
//		Country: []string{"China"},
//		Organization: []string{"Bupt"},
//		OrganizationalUnit: []string{"Bupt_Network"},
//		Locality: []string{"BeiJin Haidian XiTuChen No.10"},
//		Province: []string{"BeiJin"},
//		StreetAddress: []string{"XiTuChen Road"},
//		PostalCode: []string{"100876"},
//		SerialNumber: "10",
//		CommonName: "BUPT",
//	}
//	publicKey := model.PublicKey{
//		N: privateKey.N,
//		E: int64(privateKey.E),
//	}
//	rawCertificate := model.RawCertificate{
//		Version: 1,
//		SerialNumber: serialNumber,
//		ValidFrom: *validFrom,
//		ValidTo: *validTo,
//		Subject: subject,
//		PublicKey: publicKey,
//		Addr: model.Addr{
//			Ip: "127.0.0.1",
//			Port: 8001,
//		},
//
//	}
//	digest, err := getDigest(rawCertificate)
//	if err != nil {
//		//TODO
//		log.Println("Generate new cert failed")
//		return nil
//	}
//	var signOpts crypto.SignerOpts = crypto.SHA256
//	signature, err := privateKey.Sign(rand.Reader, digest, signOpts)
//	sig := model.Signature{
//		Subject: subject,
//		Signature: signature,
//	}
//	cert := model.Certificate{
//		RawCertificate: rawCertificate,
//		Signatures: []model.Signature{sig},
//	}
//
//	certContent, err := cert.MarshalBinary()
//	if err != nil {
//		//TODO
//		log.Println(err)
//	}
//	certBlock := &pem.Block{
//		Type: "Root Certificate",
//		Bytes: certContent,
//	}
//	certFile, err := os.Create("RootCert.pem.cert")
//	if err != nil {
//		//TODO
//		log.Println(err)
//	}
//	err = pem.Encode(certFile, certBlock)
//	return &cert
//}
//
//func loadCertificate(fileName string) (*model.Certificate, error) {
//	file, err := os.Stat(fileName)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//
//	certFile, err := os.OpenFile(file.Name(), os.O_RDONLY, 0)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	defer certFile.Close()
//
//	var content []byte = make([]byte, file.Size())
//	_, err = certFile.Read(content)
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//
//	certBlock, _ := pem.Decode(content)
//	certContent := certBlock.Bytes
//	var cert model.Certificate
//	_, err = asn1.Unmarshal(certContent, &cert)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//
//	return &cert, nil
//}
//
//func loadPrivateKey(fileName string) (*rsa.PrivateKey, error) {
//	privFileInfo, err := os.Stat(fileName)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	privFile, err := os.OpenFile(fileName, os.O_RDONLY, 0)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	defer privFile.Close()
//	content := make([]byte, privFileInfo.Size())
//	_, err = privFile.Read(content)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	priBlock, _ := pem.Decode(content)
//	var private rsa.PrivateKey
//	_, err = asn1.Unmarshal(priBlock.Bytes, private)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	return &private, nil
//}
//
//func getDigest(v interface{}) ([]byte, error) {
//	content, err := asn1.Marshal(v)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return []byte{}, err
//	}
//	hash := crypto.SHA256.New()
//	hash.Write(content)
//	digest := hash.Sum(nil)
//	return digest, nil
//}
//
//func NewCA() CAImpl {
//	return CAImpl{
//		Certificates: map[string]*model.Certificate{},
//	}
//}
//
////daemon process for dealing with client request
//func CARun() {
//
//}
//
////rpc methods begins
//
////check the certificate among all validated certs
//func validateNewCert(certificate model.Certificate) bool {
//	certName, err := certificate.GetName()
//	if err != nil {
//		//TODO
//		log.Fatal(err)
//	}
//	if oldCert, ok := CertificateAuthority.Certificates[certName]; ok {
//		//Update cert
//		if certificate.SerialNumber.Cmp(oldCert.SerialNumber) != 1 {
//			//TODO
//			log.Println("SerialNumber is invalid")
//			return false
//		}
//
//	} else {
//		//Add a new cert
//	}
//	//validate signatures
//	validF, err := ptypes.Timestamp(&certificate.ValidFrom)
//	if err != nil {
//		//TODO
//		log.Println(err)
//	}
//	if validF.After(time.Now()) {
//		//TODO
//		log.Println("ValidFrom timeStamp is wrong")
//		return false
//	}
//	validT, err := ptypes.Timestamp(&certificate.ValidTo)
//	if err != nil {
//		//TODO
//		log.Println(err)
//	}
//	if validT.Before(time.Now()) {
//		//TODO
//		log.Println("ValidTo timeStamp is wrong")
//		return false
//	}
//	digest, err := getDigest(certificate.RawCertificate)
//	if err != nil {
//		//TODO
//		log.Println("Validate cert failed")
//		return false
//	}
//	for _, subject := range certificate.Signatures {
//		subjectName, err := subject.String()
//		if err != nil {
//			log.Println(err)
//			return false
//		}
//		if cert, ok := CertificateAuthority.Certificates[subjectName]; ok {
//			err = rsa.VerifyPKCS1v15(&rsa.PublicKey{
//				N: cert.PublicKey.N,
//				E: int(cert.PublicKey.E),
//			}, crypto.SHA256, digest, subject.Signature)
//			if err != nil {
//				//TODO
//				//signature validate failed
//				log.Println(err)
//				return false
//			}
//		} else {
//			newSubject, err := certificate.Subject.String()
//			if err != nil {
//				//TODO
//				log.Println(err)
//				return false
//			}
//			if newSubject == subjectName {
//				err = rsa.VerifyPKCS1v15(&rsa.PublicKey{
//					N: certificate.PublicKey.N,
//					E: int(certificate.PublicKey.E),
//				}, crypto.SHA256, digest, subject.Signature)
//				if err != nil {
//					//TODO
//					//signature validate failed
//					log.Println(err)
//					return false
//				}
//			} else {
//				//Non-existent subject's signature
//				return false
//			}
//		}
//	}
//	return true
//}
//
//func IsOverdue(certificate model.Certificate) bool {
//	validT, err := ptypes.Timestamp(&certificate.ValidTo)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return false
//	}
//	if validT.Before(time.Now()) {
//		//TODO
//		log.Println("This cert is overdue")
//		return false
//	}
//	return true
//}
//
//type InvalidCertErr struct {
//	Msg string
//}
//
//func (in InvalidCertErr) Error() string {
//	return in.Msg
//}
//
//func (ca *CAImpl) GetCertificate(context.Context, *pb.Reserved) (*pb.Certificate, error) {
//	panic("implement me")
//}
//
//func (ca *CAImpl) ProposalCert(ctx context.Context, in *pb.Certificate) (*pb.Reserved, error) {
//	res, err := utils.Convert(*in, model.ModelCertT, true)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	cert := res.(model.Certificate)
//	if !validateNewCert(cert) {
//		log.Println("Invalid Certificate")
//		return nil, InvalidCertErr{"Invalid certificate"}
//	}
//	certName, err := cert.Subject.String()
//	if certToAudit, ok := ca.certToAudit[certName]; ok {
//		if certToAudit.SerialNumber.Cmp(cert.SerialNumber) != 1 {
//			return nil, InvalidCertErr{""}
//		}
//	}
//	ca.certToAudit[certName] = &cert
//	return nil, nil
//}
//
////Process the result of auditing certificate
//func (ca *CAImpl) AuditResponse(ctx context.Context, result *pb.AuditResult) (*pb.Reserved, error) {
//	if result.Result {
//		//cert, err := model.ProtoCertToCert(*result.Cert)
//		res, err := utils.Convert(*result.Subject, reflect.TypeOf(model.Subject{}), true)
//		if err != nil {
//			log.Println(err)
//			return nil, err
//		}
//		subject := res.(model.Subject)
//		name, err := subject.String()
//		if err != nil {
//			log.Println("Audit Response failed", err)
//			return nil, err
//		}
//		if cert, ok := ca.Certificates[name]; ok {
//			content, err := cert.RawCertificate.MarshalBinary()
//			if err != nil {
//				log.Println("Audit Response failed", err)
//				return nil, err
//			}
//			for _, sig := range cert.Signatures {
//				sigSubject, err := sig.Subject.String()
//				if err != nil {
//					log.Println("Audit Response failed", err)
//					return nil, err
//				}
//				if sigSubject == name {
//					legal, err := cert.ValidateSignature(content, sig.Signature)
//					if err != nil {
//						log.Println("Audit Response failed", err)
//						return nil, err
//					}
//					if legal {
//
//					}
//				}
//			}
//		}
//	}
//	panic("implement me")
//}
//
//func (ca *CAImpl) ShowUnauditedCert(r *pb.Reserved, stream pb.CertificateAuthority_ShowUnauditedCertServer) error {
//	for _, cert := range ca.certToAudit {
//		res, err := utils.Convert(*cert, model.PbCertT, true)
//		if err != nil {
//			log.Println(err)
//		}
//		certProto := res.(pb.Certificate)
//		if err := stream.Send(&certProto); err != nil {
//			//TODO
//			log.Println(err)
//		}
//	}
//	return nil
//}
//
//type AuditError struct {
//	Msg string
//}
//
//func (this AuditError) Error() string {
//	return this.Msg
//}
//
//func (ca *CAImpl) AuditCert(ctx context.Context, request *pb.AuditRequest) (*pb.Reserved, error) {
//	res, err := utils.Convert(*request.Cert, model.ModelCertT, true)
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	certAuit := res.(model.Certificate)
//	//check certificate
//	if len(certAuit.Signatures) != 1 {
//		log.Println("Incorrect signatures")
//		return nil, AuditError{Msg:"Incorrect signatures:too much"}
//	}
//
//	signature, err := ca.SignatureCrt(certAuit)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	certRoot, err := loadCertificate(RootPrivateKey)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	certAuit.Signatures = append(certAuit.Signatures, model.Signature{
//		Subject: certRoot.Subject,
//		Signature: signature,
//	})
//	res, err = utils.Convert(certRoot, model.PbCertT, true)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	certRootProto := res.(pb.Certificate)
//
//	//send it back to the client
//	res, err = utils.Convert(certAuit, model.PbCertT, true)
//	certAuditProto := res.(pb.Certificate)
//	conn, err := grpc.Dial(certAuit.GetAddr())
//	if err != nil {
//		//TODO
//		log.Println(err)
//		return nil, err
//	}
//	defer conn.Close()
//	client := pb.NewCertificateAuthorityClient(conn)
//	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
//	defer cancel()
//	message := pb.AuditResult{
//		Cert: &certAuditProto,
//		Result: true,
//		Subject: certRootProto.RawCertificate.Subject,
//	}
//	_, err = client.AuditResponse(ctx, &message)
//	if err != nil {
//		log.Println(err)
//		return nil, err
//	}
//	return nil, nil
//}
//
////rpc methods end
//
////network nodes cooperate
//func IssueCert() {
//	cert := CertificateAuthority.GenerateCertificate()
//	CertificateAuthority.Propagate(cert)
//}
