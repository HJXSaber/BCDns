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
	"github.com/syndtr/goleveldb/leveldb"
	"reflect"
	"sync"
	"time"
)

//Define operation types

type OperationType uint8

const (
	Add OperationType = iota
	Del
	Mod
)

const Dereliction = "No owner"

type ProposalMassage struct {
	Body      ProposalBody
	Signature []byte
}

type ProposalBody struct {
	Timestamp int64
	PId       PId
	Owner     string
	Type      OperationType
	ZoneName  string
	Nonce     uint32 //Pow hashcode
	Values    map[string]string
	Id        []byte
}

func (p *ProposalMassage) ValidateAdd() bool {
	if !service.CertificateAuthorityX509.Exits(p.Body.PId.Name) {
		fmt.Printf("[ValidateAdd] Invalid HostName=%v", p.Body.PId.Name)
		return false
	}
	bodyByte, err := p.Body.Hash()
	if err != nil {
		fmt.Printf("[ValidateAdd] json.Marshal failed err=%v\n", err)
		return false
	}
	if time.Now().Before(time.Unix(p.Body.Timestamp, 0)) {
		fmt.Printf("[ValidateAdd] TimeStamp is invalid t=%v\n", p.Body.Timestamp)
		return false
	}
	blockProposal := new(ProposalMassage)
	if data, err := dao.Dao.GetZoneName(p.Body.ZoneName); err != leveldb.ErrNotFound {
		blockProposal, err = UnMarshalProposalMassage(data)
		if err != nil {
			fmt.Printf("[ValidateAdd] UnMarshalProposalMassage error=%v\n", err)
			return false
		}
		if blockProposal.Body.Owner != Dereliction {
			fmt.Printf("[ValidateAdd] ZoneName exits or get failed err=%v\n", err)
			return false
		}
	}
	if !service.CertificateAuthorityX509.VerifySignature(p.Signature, bodyByte, p.Body.PId.Name) {
		fmt.Printf("[ValidateAdd] validate signature falied\n")
		return false
	}
	if !p.Body.ValidatePow() {
		fmt.Printf("[ValidateAdd] validate Pow faliled\n")
		return false
	}
	return true
}

func (p *ProposalMassage) ValidateDel() bool {
	if !service.CertificateAuthorityX509.Exits(p.Body.PId.Name) {
		fmt.Printf("[ValidateDel] Invalid HostName=%v", p.Body.PId.Name)
		return false
	}
	bodyByte, err := p.Body.Hash()
	if err != nil {
		fmt.Printf("[ValidateDel] json.Marshal failed err=%v\n", err)
		return false
	}
	if time.Now().Before(time.Unix(p.Body.Timestamp, 0)) {
		fmt.Printf("[ValidateDel] TimeStamp is invalid t=%v\n", p.Body.Timestamp)
		return false
	}
	if p.Body.Owner != Dereliction {
		fmt.Printf("[ValidateDel] Owner is wrong\n")
		return false
	}
	blockProposal := new(ProposalMassage)
	if data, err := dao.Dao.GetZoneName(p.Body.ZoneName); err == leveldb.ErrNotFound {
		fmt.Printf("[ValidateDel] ZoneName is not exist\n")
		return false
	} else {
		blockProposal, err = UnMarshalProposalMassage(data)
		if err != nil {
			fmt.Printf("[ValidateDel] UnMarshalProposalMassage error=%v\n", err)
			return false
		}
	}
	if p.Body.PId.Name != blockProposal.Body.Owner {
		fmt.Println("[ValidateDel] Zonename %v is not belong to %v\n", p.Body.ZoneName, p.Body.PId.Name)
		return false
	}
	if !service.CertificateAuthorityX509.VerifySignature(p.Signature, bodyByte, p.Body.PId.Name) {
		fmt.Printf("[ValidateDel] validate signature falied\n")
		return false
	}
	return true
}

func (p *ProposalMassage) ValidateMod() bool {
	if !service.CertificateAuthorityX509.Exits(p.Body.PId.Name) {
		fmt.Printf("[ValidateMod] Invalid HostName=%v", p.Body.PId.Name)
		return false
	}
	bodyByte, err := p.Body.Hash()
	if err != nil {
		fmt.Printf("[ValidateMod] json.Marshal failed err=%v\n", err)
		return false
	}
	if time.Now().Before(time.Unix(p.Body.Timestamp, 0)) {
		fmt.Printf("[ValidateMod] TimeStamp is invalid t=%v\n", p.Body.Timestamp)
		return false
	}
	blockProposal := new(ProposalMassage)
	if data, err := dao.Dao.GetZoneName(p.Body.ZoneName); err == leveldb.ErrNotFound {
		fmt.Printf("[ValidateMod] ZoneName is not exist\n")
		return false
	} else {
		blockProposal, err = UnMarshalProposalMassage(data)
		if err != nil {
			fmt.Printf("[ValidateMod] UnMarshalProposalMassage error=%v\n", err)
			return false
		}
	}
	if p.Body.PId.Name != blockProposal.Body.Owner || p.Body.Owner != blockProposal.Body.Owner {
		fmt.Printf("[ValidateMod] Zonename %v is not belong to %v\n", p.Body.ZoneName, p.Body.PId.Name)
		return false
	}
	if !service.CertificateAuthorityX509.VerifySignature(p.Signature, bodyByte, p.Body.PId.Name) {
		fmt.Printf("[ValidateMod] validate signature falied\n")
		return false
	}
	return true
}

func (p *ProposalMassage) DeepEqual(pp ProposalMassage) bool {
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

func (p *ProposalMassage) Sign() error {
	hash, err := p.Body.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		p.Signature = sig
		return nil
	}
	return errors.New("[ProposalMassage.Sign] Generate signature failed")
}

func (p *ProposalMassage) VerifySignature() bool {
	hash, err := p.Body.Hash()
	if err != nil {
		fmt.Printf("[ProposalMassage.VerifySignature] Hash error=%v\n", err)
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(p.Signature, hash, p.Body.PId.Name)
}

func (p *ProposalMassage) MarshalProposalMassage() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnMarshalProposalMassage(data []byte) (*ProposalMassage, error) {
	p := new(ProposalMassage)
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(p); err != nil {
		return nil, err
	}
	return p, nil
}

type PId struct {
	//Name is hostname
	Name           string
	//NodeId         int64
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

func NewProposal(zoneName string, t OperationType, values map[string]string) *ProposalMassage {
	var (
		err error
		pId = PId{
			SequenceNumber: xid.New().String(),
			Name:           conf.BCDnsConfig.HostName,
		}
		body ProposalBody
	)
	switch t {
	case Add:
		body = ProposalBody{
			Timestamp: time.Now().Unix(),
			PId:       pId,
			Owner:     conf.BCDnsConfig.HostName,
			ZoneName:  zoneName,
			Type:      Add,
			Nonce:     0,
			Values:    values,
		}
		err = body.GetPowHash()
		if err != nil {
			fmt.Printf("[NewProposal] GetPowHash Failed err=%v\n", err)
			return nil
		}
	case Del:
		body = ProposalBody{
			Timestamp: time.Now().Unix(),
			PId:       pId,
			Owner:     Dereliction,
			ZoneName:  zoneName,
			Type:      Del,
			Nonce:     0,
			Values:    map[string]string{},
		}
	case Mod:
		blockProposal := new(ProposalMassage)
		if data, err := dao.Dao.GetZoneName(zoneName); err == leveldb.ErrNotFound {
			fmt.Printf("[ValidateMod] ZoneName is not exist\n")
			return nil
		} else {
			blockProposal, err = UnMarshalProposalMassage(data)
			if err != nil {
				fmt.Printf("[NewProposal] UnMarshalProposalMassage error=%v\n", err)
				return nil
			}
			if blockProposal.Body.PId.Name == Dereliction {
				fmt.Printf("[ValidateMod] ZoneName is not exist\n")
				return nil
			}
		}
		body = ProposalBody{
			Timestamp: time.Now().Unix(),
			PId:       pId,
			Owner:     conf.BCDnsConfig.HostName,
			ZoneName:  zoneName,
			Type:      Mod,
			Nonce:     0,
			Values:    utils.CoverMap(blockProposal.Body.Values, values),
		}
	default:
		fmt.Println("Unknown proposal type")
		return nil
	}
	body.Id, err = body.Hash()
	if err != nil {
		fmt.Printf("[NewProposal] Hash Failed err=%v\n", err)
		return nil
	}
	msg := &ProposalMassage{
		Body: body,
	}
	err = msg.Sign()
	if err != nil {
		fmt.Printf("[NewProposal] msg.Sign error=%v\n", err)
		return nil
	}
	return msg
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
	if err = enc.Encode(p.Owner); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.Nonce); err != nil {
		fmt.Printf("[Hash] Encode failed err=%v", err)
		return nil, err
	}
	if err = enc.Encode(p.Values); err != nil {
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

type ProposalPool struct {
	Mutex sync.Mutex
	ProposalSlice
}

func (pool *ProposalPool) Len() int {
	return len(pool.ProposalSlice)
}

func (pool *ProposalPool) Exits(pm ProposalMassage) bool {
	for _, p := range pool.ProposalSlice {
		if reflect.DeepEqual(p.Signature, pm.Signature) {
			return true
		}
	}
	return false
}

func (pool *ProposalPool) AddProposal(pm ProposalMassage) {
	pool.ProposalSlice = append(pool.ProposalSlice, pm)
}

func (pool *ProposalPool) Clear() {
	pool.ProposalSlice = ProposalSlice{}
}

func (s *ProposalSlice) FindByZoneName(name string) *ProposalMassage {
	for _, p := range *s {
		if p.Body.ZoneName == name {
			return &p
		}
	}
	return nil
}

type Endorsement struct {
	ProposalHash []byte
	Auditor      string
	Signature    []byte
}

func NewProposalAuditResponse(proposal ProposalMassage) *Endorsement {
	proposalHash, err := proposal.Body.Hash()
	if err != nil {
		fmt.Printf("[NewProposalAuditResponse] generate failed err=%v\n", err)
		return nil
	}
	if sig := service.CertificateAuthorityX509.Sign(proposalHash); sig != nil {
		return &Endorsement{
			ProposalHash: proposalHash,
			Auditor:      conf.BCDnsConfig.HostName,
			Signature:    sig,
		}
	}
	fmt.Printf("[NewProposalAuditResponse] Generate signature failed")
	return nil
}

func (pr *Endorsement) Verify() bool {
	return service.CertificateAuthorityX509.VerifySignature(pr.Signature, pr.ProposalHash, pr.Auditor)
}

type ProposalAuditResponses map[string]Endorsement

func (prs *ProposalAuditResponses) Check() bool {

	return len(*prs) > 2*service.CertificateAuthorityX509.GetF()
}

type Endorsements struct {
	TermId     int64
	From       string
	Proposal   ProposalMassage
	Signatures map[string][]byte
	Signature  []byte
}

func NewAuditedProposal(p ProposalMassage, responses ProposalAuditResponses, termid int64) (*Endorsements, error) {
	var err error
	msg := Endorsements{
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

func (m Endorsements) Hash() ([]byte, error) {
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

func (m Endorsements) Sign() ([]byte, error) {
	hash, err := m.Hash()
	if err != nil {
		return nil, err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		return sig, nil
	}
	return nil, errors.New("Generate signature failed")
}

func (m Endorsements) VerifySignature() bool {
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

func (m Endorsements) VerifySignatures() bool {
	count := 0
	hash, err := m.Proposal.Body.Hash()
	if err != nil {
		fmt.Printf("[AuditedProposal.VerifySignatures] Hash error=%v\n", err)
		return false
	}
	for id, sig := range m.Signatures {
		if service.CertificateAuthorityX509.Exits(id) &&
			service.CertificateAuthorityX509.VerifySignature(sig, hash, id) {
			count++
		}
	}
	if count >= 2*service.CertificateAuthorityX509.GetF()+1 {
		return true
	}
	return false
}

type AuditedProposalPool struct {
	Mutex sync.Mutex
	AuditedProposalSlice
}

type AuditedProposalSlice []Endorsements

func (pool *AuditedProposalPool) Len() int {
	return len(pool.AuditedProposalSlice)
}

func (pool *AuditedProposalPool) Exits(pm Endorsements) bool {
	for _, p := range pool.AuditedProposalSlice {
		if reflect.DeepEqual(p.Signature, pm.Signature) {
			return true
		}
	}
	return false
}

func (pool *AuditedProposalPool) AddProposal(pm Endorsements) {
	if !pool.Exits(pm) {
		pool.AuditedProposalSlice = append(pool.AuditedProposalSlice, pm)
	}
}

func (pool *AuditedProposalPool) Clear() {
	pool.AuditedProposalSlice = AuditedProposalSlice{}
}

func (s *AuditedProposalSlice) FindByZoneName(name string) *ProposalMassage {
	for _, p := range *s {
		if p.Proposal.Body.ZoneName == name {
			return &p.Proposal
		}
	}
	return nil
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
	hash, err := p.Body.Hash()
	if err != nil {
		return nil, err
	}
	msg := ProposalResult{
		ProposalHash: hash,
		From:         conf.BCDnsConfig.HostName,
	}
	msg.Signature, err = msg.Sign()
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

type BlockCommit struct {
	From      string
	BlockHash []byte
	Signature []byte
}

func (b BlockCommit) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(b.From); err != nil {
		return nil, err
	}
	if err := enc.Encode(b.BlockHash); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b BlockCommit) Sign() ([]byte, error) {
	hash, err := b.Hash()
	if err != nil {
		return nil, err
	}

	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		return sig, nil
	}
	return nil, errors.New("[BlockCommit] Get Signature failed")
}

func (b BlockCommit) VerifySignature() bool {
	hash, err := b.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(b.Signature, hash, b.From)
}