package messages

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/dao"
	"BCDns_0.1/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/xid"
	"math"
	"reflect"
	"time"
)

//Define operation types

type OperationType uint8

const (
	Add OperationType = iota
	Del
)

var (
	AddReqFailedType = reflect.TypeOf(AddReqFailed{})
)

type ProposalMassage struct {
	Body      ProposalBody
	Signature []byte
}

type ProposalBody struct {
	Timestamp int64
	PId       PId
	Type      OperationType
	ZoneName  string
	Nonce     uint32 //Pow hashcode
}

//type ProposalResult struct {
//	Body ResultBody
//	Sig  []byte
//}
//
//type ResultBody struct {
//	ProposalMassage
//	Result bool
//}
//
//type ProposalDealFailed struct {
//	Msg string
//}
//
//func (err ProposalDealFailed) Error() string {
//	return err.Msg
//}
//
//func (*ProposalMassage) Marshal() []byte {
//	panic("implement me")
//}
//
//func (p *ProposalMassage) Response(pass bool) ([]byte, error) {
//	body := ResultBody{
//		ProposalMassage: *p,
//		Result:          pass,
//	}
//	data, err := json.Marshal(body)
//	if err != nil {
//		return nil, err
//	}
//	sig := service.CertificateAuthorityX509.Sign(data)
//	if sig == nil {
//		return nil, ProposalDealFailed{"Sign failed"}
//	}
//	var msg ProposalResult = ProposalResult{
//		Body: body,
//		Sig:  sig,
//	}
//	data, err = json.Marshal(msg)
//	if err != nil {
//		return nil, err
//	}
//	return data, nil
//}

func (p ProposalMassage) ValidateAdd() bool {
	if !service.CertificateAuthorityX509.Exits(p.Body.PId.Name) {
		fmt.Printf("[Validate] Invalid HostName=%v", p.Body.PId.Name)
		return false
	}
	bodyByte, err := json.Marshal(p.Body)
	if err != nil {
		fmt.Printf("[Validate] json.Marshal failed err=%v\n", err)
		return false
	}
	if time.Now().Before(time.Unix(p.Body.Timestamp, 0)) {
		fmt.Printf("[Validate] TimeStamp is invalid t=%v\n", p.Body.Timestamp)
		return false
	}
	if exit, err := dao.Dao.Has([]byte(p.Body.ZoneName)); exit {
		fmt.Printf("[Validate] ZoneName exits\n")
		return false
	} else if err != nil {
		fmt.Printf("[Validate] db Has failed err=%v\n", err)
		return false
	}
	if !service.CertificateAuthorityX509.VerifySignature(p.Signature, bodyByte, p.Body.PId.Name) {
		fmt.Printf("[Validate] validate signature falied\n")
		return false
	}
	if !p.Body.ValidatePow() {
		fmt.Printf("[Validate] validate Pow faliled\n")
		return false
	}
	return true
}

func (p ProposalMassage) DeepEqual(pp ProposalMassage) bool {
	h1, err := p.Body.Hash()
	if err != nil {
		return false
	}
	h2, err := pp.Body.Hash()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(h1, h2)
}

type PId struct {
	//Name is hostname
	Name           string
	NodeId         int64
	SequenceNumber string
}

func (p PId) String() string {
	return p.Name + ":" + string(p.SequenceNumber)
}

type Operation struct {
	Type int
	//json data. Deal the data by Type
	Data []byte
}

type ProposalFunc interface {
	Do() error
	Marshal() []byte
	GetIssuer() string
	Response() ([]byte, error)
}

type DelMsg struct {
	ZoneName string
	Sig      []byte
}

func Parse(data []byte) *ProposalMassage {
	var msg ProposalMassage
	if err := json.Unmarshal(data, msg); err != nil {
		fmt.Println("Parse proposal massage failed", err)
		return nil
	}
	return &msg
}

func NewProposal(zoneName string, t OperationType) *ProposalMassage {
	var err error
	switch t {
	case Add:
		pId := PId{
			SequenceNumber: xid.New().String(),
			Name:           conf.BCDnsConfig.HostName,
		}
		msg := ProposalBody{
			Timestamp: time.Now().Unix(),
			PId:       pId,
			ZoneName:  zoneName,
			Type:      Add,
		}
		msg.HashCode, err = msg.GetPowHash()
		if err != nil {
			fmt.Printf("[NewProposal] GetPowHash Failed err=%v\n", err)
			return nil
		}
		msgByte, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("[NewProposal] json.Marshal failed err=%v", err)
			return nil
		}
		if sig := service.CertificateAuthorityX509.Sign(msgByte); sig != nil {
			return &ProposalMassage{
				Body:      msg,
				Signature: sig,
			}
		}
		return nil
	case Del:
		msg := ProposalBody{
			Timestamp: time.Now().Unix(),
			ZoneName:  zoneName,
		}
		msgByte, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("[NewProposal] json.Marshal failed err=%v", err)
			return nil
		}
		sig := service.CertificateAuthorityX509.Sign(msgByte)
		if sig == nil {
			fmt.Println("Generate proposal failed: sign failed")
			return nil
		}
		return &ProposalMassage{
			Body:      msg,
			Signature: sig,
		}
	default:
		fmt.Println("Unknown proposal type")
		return nil
	}
}

type AddReqFailed struct {
	Msg string
}

func (e AddReqFailed) Error() string {
	return e.Msg
}

type DelReqFailed struct {
	Msg string
}

func (err DelReqFailed) Error() string {
	return err.Msg
}

func doDel(data []byte, id string) error {
	var msg DelMsg
	err := json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	ok, err := dao.Dao.Has([]byte(msg.ZoneName))
	if err != nil {
		return err
	}
	if !ok {
		return DelReqFailed{"Domain name is not exited"}
	}
	if !service.CertificateAuthorityX509.VerifySignature(msg.Sig, []byte(msg.ZoneName), id) {
		return DelReqFailed{"Signature is invalid"}
	}
	return nil
}

func (p *ProposalBody) Hash() ([]byte, error) {
	var err error
	hash := sha256.New()
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err = enc.Encode(p.Timestamp); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.ZoneName); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.Type); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.PId); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.Nonce); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	hash.Write(buf.Bytes())
	return hash.Sum(nil), nil
}

func (p *ProposalBody) GetPowHash() error {
	for {
		hash, err := p.Hash()
		if err != nil {
			return err
		}
		if utils.CheckProofOfWork(utils.ProposalPOW, hash) {
			break
		} else {
			p.Nonce++
		}
	}
	return nil
}

func (p *ProposalBody) ValidatePow() bool {
	hash, err := p.Hash()
	if err != nil {
		return false
	}
	if utils.CheckProofOfWork(utils.ProposalPOW, hash) {
		return true
	}
	return false
}

type ProposalSlice []ProposalMassage

func (slice ProposalSlice) Len() int {
	return len(slice)
}

func (slice ProposalSlice) Exits(pm ProposalMassage) bool {
	for _, p := range slice {
		if reflect.DeepEqual(p.Signature, pm.Signature) {
			return true
		}
	}
	return false
}

func (slice ProposalSlice) AddProposal(pm ProposalMassage) ProposalSlice {
	// Inserted sorted by timestamp
	for i, p := range slice {
		if p.Body.Timestamp >= pm.Body.Timestamp {
			return append(append(slice[:i], pm), slice[i:]...)
		}
	}

	return append(slice, pm)
}

type ProposalAuditResponse struct {
	ProposalHash []byte
	Auditor      string
	Signature    []byte
}

func NewProposalAuditResponse(proposal ProposalMassage) *ProposalAuditResponse {
	proposalHash, err := proposal.Body.Hash()
	if err != nil {
		fmt.Printf("[NewProposalAuditResponse] generate failed err=%v\n", err)
		return nil
	}
	if sig := service.CertificateAuthorityX509.Sign(proposalHash); sig != nil {
		return &ProposalAuditResponse{
			ProposalHash: proposalHash,
			Auditor:      conf.BCDnsConfig.HostName,
			Signature:    sig,
		}
	}
	fmt.Printf("[NewProposalAuditResponse] Generate signature failed")
	return nil
}

func (pr *ProposalAuditResponse) Verify() bool {
	return service.CertificateAuthorityX509.VerifySignature(pr.Signature, pr.ProposalHash, pr.Auditor)
}

type ProposalAuditResponses map[string]ProposalAuditResponse

func (prs *ProposalAuditResponses) Check() bool {

	return len(*prs) > 2*service.CertificateAuthorityX509.GetF()
}

type AuditedProposal struct {
	TermId     int64
	From       string
	Proposal   ProposalMassage
	Signatures map[string][]byte
	Signature  []byte
}

func NewAuditedProposal(p ProposalMassage, responses ProposalAuditResponses, termid int64) (*AuditedProposal, error) {
	var err error
	msg := AuditedProposal{
		TermId:     termid,
		From:       conf.BCDnsConfig.HostName,
		Proposal:   p,
		Signatures: map[string][]byte{},
	}
	for id, r := range responses {
		msg.Signatures[id] = r.Signature
	}
	msg.Signature, err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (m AuditedProposal) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(m.TermId); err != nil {
		return nil, err
	}
	if err := enc.Encode(m.From); err != nil {
		return nil, err
	}
	if err := enc.Encode(m.Proposal); err != nil {
		return nil, err
	}
	if err := enc.Encode(m.Signatures); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m AuditedProposal) Sign() ([]byte, error) {
	hash, err := m.Hash()
	if err != nil {
		return nil, err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		return sig, nil
	}
	return nil, errors.New("Generate signature failed")
}

func (m AuditedProposal) VerifySignature() bool {
	hash, err := m.Hash()
	if err != nil {
		return false
	}
	if service.CertificateAuthorityX509.Exits(m.From) &&
		service.CertificateAuthorityX509.VerifySignature(m.Signature, hash, m.From) {
		return true
	}
	return false
}

func (m AuditedProposal) VerifySignatures() bool {
	count := 0
	for id, sig := range m.Signatures {
		if service.CertificateAuthorityX509.Exits(id) &&
			service.CertificateAuthorityX509.VerifySignature(sig, m.Proposal.Body.HashCode, id) {
			count++
		}
	}
	if count >= 2*service.CertificateAuthorityX509.GetF()+1 {
		return true
	}
	return false
}

type ProposalResult struct {
	ProposalHash []byte
	From         string
	Signature    []byte
}

func (p ProposalResult) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(p.ProposalHash); err != nil {
		return nil, err
	}
	if err := enc.Encode(p.From); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p ProposalResult) Sign() ([]byte, error) {
	hash, err := p.Hash()
	if err != nil {
		return nil, err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		return sig, nil
	}
	return nil, errors.New("Generate ProposalResult Signature failed")
}

func (p ProposalResult) VerifySignature() bool {
	hash, err := p.Hash()
	if err != nil {
		fmt.Printf("[ProposalResult.VerifySignature] Generate Hash failed\n")
		return false
	}
	if service.CertificateAuthorityX509.VerifySignature(p.Signature, hash, p.From) {
		return true
	}
	return false
}

func NewProposalResult(p ProposalMassage) (*ProposalResult, error) {
	var err error
	msg := ProposalResult{
		ProposalHash: p.Body.HashCode,
		From:         conf.BCDnsConfig.HostName,
	}
	msg.Signature, err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
