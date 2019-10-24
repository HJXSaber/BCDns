package service

import (
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
)

type Node struct {
}

type NodeInterface interface {
	Run()
}

func (n Node) Run() {
	for true {
		select {
		case proposalByte := <-service.ProposalChan:
			var proposal messages.ProposalMassage
			err := json.Unmarshal(proposalByte, &proposal)
			if err != nil {
				fmt.Printf("[Node.Run] json.Unmarshal failed err=%v\n", err)
				continue
			}
			switch proposal.Body.Type {
			case messages.Add:
				go handleAddProposal(proposal)
			case messages.Del:
				go handleDelProposal(proposal)
			}
		}
	}
}

func handleAddProposal(proposal messages.ProposalMassage) {
	if !proposal.ValidateAdd() {
		return
	}
	if auditResponse := messages.NewProposalAuditResponse(proposal); auditResponse != nil {
		auditResponseByte, err := json.Marshal(auditResponse)
		if err != nil {
			fmt.Printf("[handleAddProposal] json.Marshal failed err=%v\n", err)
			return
		}
		service.P2PNet.SendTo(auditResponseByte, service.AuditResponse, proposal.Body.PId.Name)
	}
}

func handleDelProposal(proposal messages.ProposalMassage) {

}
