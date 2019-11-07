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
			switch proposal.Body.Type {
			case messages.Add:
				go handleAddProposal(&proposal)
			case messages.Del:
				go handleDelProposal(proposal)
			}
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
		fmt.Println("auditedR", auditResponse)
		service.P2PNet.SendTo(auditResponseByte, service.AuditResponse, proposal.Body.PId.Name)
		blockChain.NodeProposalPool.AddProposal(*proposal)
	}
}

func handleDelProposal(proposal messages.ProposalMassage) {

}

func ProcessBlock(block *blockChain.Block) {
	proposalsPool := new(messages.ProposalPool)
	for _, p := range block.ProposalSlice {
		if blockChain.NodeProposalPool.Exits(p) {
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
			service.P2PNet.SendTo(msgByte, service.ProposalResult, p.Body.PId.Name)
			proposalsPool.AddProposal(p)
		}
	}
	newBlock, err := blockChain.BlockChain.MineBlock(proposalsPool.ProposalSlice)
	if err != nil {
		fmt.Printf("[ProcessBlock] error=%v\n", err)
		return
	}
	err = blockChain.BlockChain.AddBlock(newBlock)
	if err != nil {
		fmt.Printf("[ProcessBlock] error=%v\n", err)
		return
	}
	blockChain.NodeProposalPool.Clear()
}

func NewNode() *NodeT {
	return &NodeT{}
}
