package service

import (
	"BCDns_0.1/blockChain"
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
				go handleAddProposal(&proposal)
			case messages.Del:
				go handleDelProposal(proposal)
			}
		case msgByte := <-service.BlockChan:
			block := new(blockChain.Block)
			if err := block.UnmarshalBinary(msgByte); err != nil {
				fmt.Printf("[Node.RUn] block.UnmarshalBinary failed err=%v\n", err)
				continue
			}
			if !block.VerifyBlock() {
				fmt.Printf("[Node.Run] Verify block failed block=%v\n", block)
				continue
			}
			ProcessBlock(block)
		}
	}
}

func handleAddProposal(proposal *messages.ProposalMassage) {
	if !proposal.ValidateAdd() {
		return
	}
	if auditResponse := messages.NewProposalAuditResponse(*proposal); auditResponse != nil {
		auditResponseByte, err := json.Marshal(auditResponse)
		if err != nil {
			fmt.Printf("[handleAddProposal] json.Marshal failed err=%v\n", err)
			return
		}
		service.P2PNet.SendTo(auditResponseByte, service.AuditResponse, proposal.Body.PId.NodeId)
		blockChain.Bl.CurrentBlock.AddProposal(&proposal)
	}
}

func handleDelProposal(proposal messages.ProposalMassage) {

}

func ProcessBlock(block *blockChain.Block) {
	b := blockChain.NewBlock(blockChain.Bl.PreviousBlock().Hash())
	for _, p := range *block.ProposalSlice {
		if blockChain.Bl.CurrentBlock.Exits(p) {
			msg, err := messages.NewProposalResult(p)
			if err != nil {
				fmt.Printf("[ProcessBlock] Generate proposalResult failed err=%v\n", err)
				continue
			}
			msgByte, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("[ProcessBlock] json.Marshal failed err=%v\n", err)
				continue
			}
			service.P2PNet.SendTo(msgByte, service.ProposalResult, p.Body.PId.NodeId)
			b.AddProposal(&p)
		}
	}
	blockChain.Bl.AddBlock(b)
}
