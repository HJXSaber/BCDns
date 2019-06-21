package messages

import (
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/dao"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
)

//Define operation types
const (
	Add = iota
	Del
)

type ProposalMassage struct {
	PId
	Operation
}

type ProposalResult struct {

}

func (p *ProposalMassage) Do() error {
	switch p.Type {
	case Add:
		if err := doAdd(p.data, p.GetIssuer()); err != nil {
			fmt.Println("Process proposal failed", err)

			return err
		}
	case Del:

	default:
		
	}
	return nil
}

func (*ProposalMassage) GetIssuer() string {

	panic("")
}

func (*ProposalMassage) Marshal() []byte {
	panic("implement me")
}

type PId struct {
	Name string
	SequenceNumber uuid.UUID
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
}

type AddMsg struct {
	ZoneName string
	Sig []byte
}

func Parse([]byte) *ProposalMassage {
	return nil
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