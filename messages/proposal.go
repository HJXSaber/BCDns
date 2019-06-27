package messages

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/dao"
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"reflect"
)

//Define operation types
const (
	Add = iota
	Del
)

var (
	AddReqFailedType = reflect.TypeOf(AddReqFailed{})
)

type ProposalMassage struct {
	PId
	Operation
}

type ProposalResult struct {
	Body ResultBody
	Sig []byte
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

func (p *ProposalMassage) Do() error {
	switch p.Type {
	case Add:
		if err := doAdd(p.data, p.GetIssuer()); err != nil {
			fmt.Println("Process proposal failed", err)
			return err
		}
	case Del:
		if err := doDel(p.data, p.GetIssuer()); err != nil {
			fmt.Println("Process proposal failed", err)
		}
	default:
		return ProposalDealFailed{"Do: Unknown proposal massage type"}
		
	}
	return nil
}

func (p *ProposalMassage) GetIssuer() string {
	return p.Name
}

func (*ProposalMassage) Marshal() []byte {
	panic("implement me")
}

func (p *ProposalMassage) Response(pass bool) ([]byte, error) {
	body := ResultBody{
		ProposalMassage: *p,
		Result: pass,
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
		Sig: sig,
	}
	data, err = json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type PId struct {
	//Name is hostname
	Name string
	SequenceNumber string
}


func (p PId) String() string {
	return p.Name + ":" + string(p.SequenceNumber)
}

type Operation struct {
	Type int
	//json data. Deal the data by Type
	data []byte
}

type ProposalFunc interface {
	Do() error
	Marshal() []byte
	GetIssuer() string
	Response() ([]byte, error)
}

type AddMsg struct {
	ZoneName string
	Sig []byte
}

type DelMsg struct {
	ZoneName string
	Sig []byte
}

func Parse(data []byte) *ProposalMassage {
	var msg ProposalMassage
	if err := json.Unmarshal(data, msg); err != nil {
		fmt.Println("Parse proposal massage failed", err)
		return nil
	}
	return &msg
}

func NewProposal(zoneName string, t int) *ProposalMassage {
	switch t {
	case Add:
		sig := service.CertificateAuthorityX509.Sign([]byte(zoneName))
		if sig == nil {
			fmt.Println("Generate proposal failed: sign failed")
			return nil
		}
		msg := AddMsg{
			ZoneName:zoneName,
			Sig:sig,
		}
		msgData, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return &ProposalMassage{
			PId: PId{
				Name: conf.BCDnsConfig.HostName,
				SequenceNumber: xid.New().String(),
			},
			Operation: Operation{
				Type: Add,
				data: msgData,
			},
		}
	case Del:
		sig := service.CertificateAuthorityX509.Sign([]byte(zoneName))
		if sig == nil {
			fmt.Println("Generate proposal failed: sign failed")
			return nil
		}
		msg := DelMsg{
			ZoneName:zoneName,
			Sig:sig,
		}
		msgData, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return &ProposalMassage{
			PId: PId{
				Name: conf.BCDnsConfig.HostName,
				SequenceNumber: xid.New().String(),
			},
			Operation: Operation{
				Type: Del,
				data: msgData,
			},
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

func doAdd(data []byte, id string) error {
	var msg AddMsg
	err := json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	ok, err := dao.Dao.Has([]byte(msg.ZoneName))
	if err != nil {
		return err
	}
	if ok {
		return AddReqFailed{"Domain name is occupied"}
	}
	if !service.CertificateAuthorityX509.VerifySignature(msg.Sig, []byte(msg.ZoneName), id) {
		return AddReqFailed{"Signature is invalid"}
	}
	return nil
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
