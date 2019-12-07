package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/network/service"
	"encoding/json"
)

const (
	keep uint8 = iota
	drop
)

type Node struct {
	Proposals       map[string]uint8
	Blocks          map[string]blockChain.BlockMessage
	BlockPrepareMsg map[string]map[string][]byte
}

func (n *Node) Run(done chan uint8) {
	defer close(done)
	for {
		select {
		case msgByte := <-service.ProposalChan:
			if service.Leader.IsLeader() {
				ProposalMessageChan <- msgByte
			} else {
				n.handleProposal(msgByte)
			}
		case msgByte := <-service.BlockChan:
			var msg blockChain.BlockMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			id, err := msg.Block.Hash()
			if err != nil {
				logger.Warningf("[Node.Run] block.Hash error=%v", err)
				continue
			}
			if !ValidateBlock(&msg) {
				continue
			}
			n.Blocks[string(id)] = msg
			n.BlockPrepareMsg[string(id)] = map[string][]byte{}
			blockConfirmMsg, err := NewBlockConfirmMessage(id)
			if err != nil {
				logger.Warningf("[Node.Run] NewBlockConfirmMessage error=%v", err)
				continue
			}
			jsonData, err := json.Marshal(blockConfirmMsg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Marshal error=%v", err)
				continue
			}
			service.Net.BroadCast(jsonData, service.BlockConfirmMsg)
		case msgByte := <-service.BlockConfirmChan:
			var msg BlockConfirmMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if !msg.Verify() {
				logger.Warningf("[Node.Run] msg.Verify failed")
				continue
			}
			n.BlockPrepareMsg[string(msg.Id)][msg.From] = msg.Signature
			if service2.CertificateAuthorityX509.Check(len(n.BlockPrepareMsg[string(msg.Id)][msg.From])) {

			}
		}
	}
}

func (n *Node) handleProposal(msgByte []byte) {
	var proposal ProposalMessage
	err := json.Unmarshal(msgByte, &proposal)
	if err != nil {
		logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
		return
	}
	switch proposal.Type {
	case Add:
		if !proposal.ValidateAdd() {
			logger.Warningf("[handleProposal] ValidateAdd failed")
			return
		}
	case Del:
		if !proposal.ValidateDel() {
			logger.Warningf("[handleProposal] ValidateDel failed")
			return
		}
	case Mod:
		if !proposal.ValidateMod() {
			logger.Warningf("[handleProposal] ValidateMod failed")
			return
		}
	}
	n.Proposals[string(proposal.Id)] = keep
	service.Net.SendToLeader(msgByte, service.ProposalMsg)
}

func ValidateBlock(msg *blockChain.BlockMessage) bool {
	if !service2.CertificateAuthorityX509.Exits(msg.From) {
		logger.Warningf("[ValidateBlock] msg.From is not exist")
		return false
	}
	if !msg.VerifyBlock() {
		logger.Warningf("[ValidateBlock] VerifyBlock failed")
		return false
	}
	if !msg.VerifySignature() {
		logger.Warningf("[ValidateBlock] VerifySignature failed")
		return false
	}
	if !ValidateProposals(msg) {
		logger.Warningf("[ValidateBlock] ValidateProposals failed")
		return false
	}
	return true
}
