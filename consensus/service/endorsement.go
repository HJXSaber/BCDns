package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/messages"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"
)

var (
	Endorsement *EndorsementT
)

type EndorsementT struct {
	ProposalChan chan messages.ProposalMassage
	ProposalToAudit chan []byte
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
		ProposalChan: make(chan messages.ProposalMassage, conf.BCDnsConfig.ProposalBufferSize),
		ProposalToAudit: make(chan []byte, conf.BCDnsConfig.ProposalBufferSize),
		Responses: make(map[messages.PId]Proposal),
	}
}

//Collect endorsement
func (endorsement *EndorsementT) ProcessProposal() {
	for {
		msg := <- endorsement.ProposalChan
		if _, ok := endorsement.Responses[msg.PId]; ok {
			fmt.Printf("Proposal %s exits", msg.PId)
			continue
		}
		proposal := Proposal{
			Type:conf.ProposalMsg,
			Msg:msg,
			//Timer is set to clean overtime proposal massage
			Timer:time.AfterFunc(conf.BCDnsConfig.ProposalOvertime, func() {
				delete(endorsement.Responses, msg.PId)
			}),

		}
		endorsement.Responses[msg.PId] = proposal
	}
}

func (endorsement *EndorsementT) PutProposal(massage messages.ProposalMassage) {
	endorsement.ProposalChan <- massage
}

func (endorsement *EndorsementT) AuditProposal() {
	var (
		msg messages.ProposalMassage
	)
	for {
		msgByte := <- endorsement.ProposalToAudit
		err := json.Unmarshal(msgByte, msg)
		if err != nil {
			fmt.Println("Audit proposal failed", err)
			continue
		}
		switch msg.Type {
		case messages.Add:
			var addMsg messages.AddMsg
			err := json.Unmarshal(msg.Data, addMsg)
			if err != nil {
				fmt.Println("Audit proposal failed", err)
				continue
			}

		case messages.Del:
		default:
			fmt.Println("Audit proposal: unknown operation type")
			continue
		}
	}
}

type EndorsementInterface interface {
	PutProposal(massage messages.ProposalMassage)
	ProcessProposal()
	AuditProposal()
}

