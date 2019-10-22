package service

import (
	"BCDns_0.1/bcDns/conf"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/dao"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"time"
)

var (
	Endorsement *EndorsementT
)

type EndorsementT struct {
	ProposalChan    chan messages.ProposalMassage
	ProposalToAudit chan []byte
	Responses       map[messages.PId]Proposal
}

type Proposal struct {
	Type  uint8
	Msg   messages.ProposalMassage
	Timer *time.Timer
	Sigs  [][]byte
}

func init() {
	Endorsement = &EndorsementT{
		ProposalChan:    make(chan messages.ProposalMassage, conf.BCDnsConfig.ProposalBufferSize),
		ProposalToAudit: make(chan []byte, conf.BCDnsConfig.ProposalBufferSize),
		Responses:       make(map[messages.PId]Proposal),
	}
}

//Collect endorsement
func (endorsement *EndorsementT) ProcessProposal() {
	for {
		msg := <-endorsement.ProposalChan
		msgByte, err := json.Marshal(msg)
		if err != nil {
			fmt.Print("Process Proposal failed", err)
			continue
		}
		if _, ok := endorsement.Responses[msg.Body.PId]; ok {
			fmt.Printf("Proposal %s exits", msg.Body.PId)
			continue
		}
		proposal := Proposal{
			Type: conf.ProposalMsg,
			Msg:  msg,
			//Timer is set to clean overtime proposal massage
			Timer: time.AfterFunc(conf.BCDnsConfig.ProposalOvertime, func() {
				delete(endorsement.Responses, msg.Body.PId)
			}),
		}
		endorsement.Responses[msg.Body.PId] = proposal
		service.P2PNet.BroadcastMsg(msgByte)
	}
}

func (endorsement *EndorsementT) PutProposal(massage messages.ProposalMassage) {
	endorsement.ProposalChan <- massage
}

func (endorsement *EndorsementT) EnqueueAuditProposal() {
	var (
		msg messages.ProposalMassage
	)
	for {
		msgByte := <-endorsement.ProposalToAudit
		err := json.Unmarshal(msgByte, msg)
		if err != nil {
			fmt.Println("Audit proposal failed", err)
			continue
		}
		switch msg.Body.Type {
		case messages.Add:
			err = dao.Dao.Add(msg.Body.HashCode, msgByte)
			if err != nil {
				fmt.Printf("[EnqueueAuditProposal] Add msg failed err=%v", err)
				return
			}
		case messages.Del:
			bodyByte, err := json.Marshal(msg.Body)
			if err != nil {
				fmt.Printf("[EnqueueAuditProposal] json.Marshal failed err=%v", err)
				return
			}
			if service2.CertificateAuthorityX509.VerifySignature(msg.Signature, bodyByte, msg.Body.PId.Name) {
				if err := dao.Dao.DelEX(msg.Body.HashCode); err != nil {
					fmt.Printf("[EnqueueAuditProposal] Del dns record failed err=%v", err)
					return
				}
			} else {
				fmt.Printf("[EnqueueAuditProposal] Wrong request msg=%v", msg)
				return
			}
		default:
			fmt.Println("Audit proposal: unknown operation type")
			continue
		}
	}
}

type EndorsementInterface interface {
	PutProposal(massage messages.ProposalMassage)
	//Deal the proposal intern
	ProcessProposal()
	//Deal the proposal send by other peer
	EnqueueAuditProposal()
}
