package messages

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/dao"
	"crypto/sha1"
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
	HashCode  []byte
}

type ProposalResult struct {
	Body ResultBody
	Sig  []byte
}

type ResultBody struct {
	ProposalMassage
	Result bool
}

type ProposalDealFailed struct {
	Msg string
}

func (err ProposalDealFailed) Error() string {
	return err.Msg
}

func (*ProposalMassage) Marshal() []byte {
	panic("implement me")
}

func (p *ProposalMassage) Response(pass bool) ([]byte, error) {
	body := ResultBody{
		ProposalMassage: *p,
		Result:          pass,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	sig := service.CertificateAuthorityX509.Sign(data)
	if sig == nil {
		return nil, ProposalDealFailed{"Sign failed"}
	}
	var msg ProposalResult = ProposalResult{
		Body: body,
		Sig:  sig,
	}
	data, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type PId struct {
	//Name is hostname
	Name           string
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
	switch t {
	case Add:
		pId := PId{
			SequenceNumber: xid.New().String(),
			Name:           conf.BCDnsConfig.HostName,
		}
		hashCode, err := getHashCode(pId, zoneName, Add)
		if err != nil {
			fmt.Printf("[NewProposal] err=%v", err)
			return nil
		}
		msg := ProposalBody{
			Timestamp: time.Now().Unix(),
			PId:       pId,
			ZoneName:  zoneName,
			Type:      Add,
			HashCode:  hashCode,
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

func getHashCode(pId PId, zoneName string, operationType OperationType) ([]byte, error) {
	hash := sha1.New()
	pIdByte, err := json.Marshal(pId)
	if err != nil {
		fmt.Printf("[getHashCode] json.Marshal failed err=%v", err)
		return nil, err
	}
	hash.Write(pIdByte)
	hash.Write([]byte(zoneName))
	hash.Write([]byte{byte(operationType)})
	sum1 := hash.Sum(nil)
	for i := 1; i < math.MaxInt64; i++ {
		hash.Reset()
		hash.Write([]byte{byte(i)})
		sum2 := hash.Sum(nil)
		count := 0
		for i, v := range sum1 {
			if sum2[i]+v == uint8(0) {
				count++
				if count >= conf.BCDnsConfig.PowDifficult {
					return sum2, nil
				}
			}
		}
	}
	return nil, errors.New("[getHashCode]Cannot find appropriate value")
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

func (slice ProposalSlice) AddTransaction(pm ProposalMassage) ProposalSlice {
	// Inserted sorted by timestamp
	for i, p := range slice {
		if p.Body.Timestamp >= pm.Body.Timestamp {
			return append(append(slice[:i], pm), slice[i:]...)
		}
	}

	return append(slice, pm)
}
