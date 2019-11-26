package service

import (
	"BCDns_0.1/blockChain"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
)

var (
	Node *NodeT
)

type NodeT struct {
}

type NodeInterface interface {
	Run()
}

func (n NodeT) Run(done chan uint) {
	defer close(done)
	for true {
		select {
		case proposalByte := <-service.ProposalChan:
			var proposal messages.ProposalMassage
			err := json.Unmarshal(proposalByte, &proposal)
			if err != nil {
				fmt.Printf("[Node.Run] json.Unmarshal failed err=%v\n", err)
				continue
			}
			handleProposal(&proposal)
		case msgByte := <-service.BlockChan:
			blockMsg := new(blockChain.BlockMessage)
			err := json.Unmarshal(msgByte, blockMsg)
			if err != nil {
				fmt.Printf("[Node.RUn] json.Unmarshal failed err=%v\n", err)
				continue
			}
			if !blockMsg.VerifySignature() {
				fmt.Printf("[Node.Run] VerifySignature failed msg=%v\n", blockMsg)
				continue
			}
			if !blockMsg.Block.VerifyBlock() {
				fmt.Printf("[Node.Run] Verify block failed block=%v\n", blockMsg.Block)
				continue
			}
			ProcessBlock(&blockMsg.Block)
		}
	}
}

func handleProposal(proposal *messages.ProposalMassage) {
	switch proposal.Body.Type {
	case messages.Add:
		if !proposal.ValidateAdd() {
			return
		}
	case messages.Del:
		if !proposal.ValidateDel() {
			return
		}
	case messages.Mod:
		if !proposal.ValidateMod() {
			return
		}
	default:
		fmt.Println("[handleProposal] Unknown proposal type")
	}
	if auditResponse := messages.NewProposalAuditResponse(*proposal); auditResponse != nil {
		auditResponseByte, err := json.Marshal(auditResponse)
		if err != nil {
			fmt.Printf("[handleAddProposal] json.Marshal failed err=%v\n", err)
			return
		}
		service.Net.SendTo(auditResponseByte, service.Endorsement, proposal.Body.PId.Name)
	}
}

func ProcessBlock(block *blockChain.Block) {
	err := blockChain.BlockChain.AddBlock(block)
	if err != nil {
		fmt.Printf("[ProcessBlock] error=%v\n", err)
		return
	}
	for _, p := range block.AuditedProposalSlice {
		msg, err := messages.NewProposalResult(p.Proposal)
		if err != nil {
			fmt.Printf("[ProcessBlock] Generate proposalResult failed err=%v\n", err)
			continue
		}
		msgByte, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("[ProcessBlock] json.Marshal failed err=%v\n", err)
			continue
		}
		service.Net.SendTo(msgByte, service.ProposalResult, p.Proposal.Body.PId.Name)
	}
}

func NewNode() *NodeT {
	return &NodeT{}
}
