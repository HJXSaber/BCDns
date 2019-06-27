package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"encoding/json"
	"fmt"
	"time"
)

var (
	Endorsement *EndorsementT
)

type EndorsementT struct {
	ProposalChan chan []byte
	Responses map[messages.PId]Proposal
}

type Proposal struct {
	Type uint8
	Msg messages.ProposalMassage
	Timer *time.Timer
	Sigs [][]byte
}

func init() {
	Endorsement = &EndorsementT{
		ProposalChan: make(chan []byte, conf.BCDnsConfig.ProposalBufferSize),
		Responses: make(map[messages.PId]Proposal),
	}
}

//Collect endorsement
func (endorsement *EndorsementT) ProcessProposal() {
	var msg messages.ProposalMassage
	for {
		msgByte := <- endorsement.ProposalChan
		err := json.Unmarshal(msgByte, msg)
		if err != nil {
			fmt.Println("Process proposal failed", err)
			continue
		}
		if _, ok := endorsement.Responses[msg.PId]; ok {
			fmt.Printf("Proposal %s exits", msg.PId)
			continue
		}
		proposal := Proposal{
			Msg:msg,
			Timer:time.AfterFunc(conf.BCDnsConfig.ProposalOvertime, func() {
				if len(endorsement.Responses[msg.PId].Sigs) <=
					(len(service.CertificateAuthorityX509.Certificates) - 1) / 3 {
					fmt.Println("Endorsement is not enough:", msg.PId)
					delete(endorsement.Responses, msg.PId)
				} else {

				}
			}),
		}
		endorsement.Responses[msg.PId] = proposal
	}
}

func (endorsement *EndorsementT) PutProposal(massage []byte) {
	endorsement.ProposalChan <- massage
}

type EndorsementInterface interface {
	PutProposal(massage messages.ProposalMassage)
	ProcessProposal()
}

