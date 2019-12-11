package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"bytes"
	"encoding/json"
)

const (
	unconfirmed uint8 = iota
	keep
	drop
)

const (
	ok uint8 = iota
	dataSync
	invalid
)

type Node struct {
	Proposals       map[string]uint8
	Blocks          []blockChain.BlockValidated
	Block           blockChain.Block
	BlockPrepareMsg map[string][]byte
}

func NewNode() *Node {
	return &Node{
		Proposals:       map[string]uint8{},
		Blocks:          []blockChain.BlockValidated{},
		BlockPrepareMsg: map[string][]byte{},
	}
}

func (n *Node) Run(done chan uint) {
	defer close(done)
	for {
		select {
		case msgByte := <-service.ProposalChan:
			var proposal messages.ProposalMessage
			err := json.Unmarshal(msgByte, &proposal)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if !n.handleProposal(proposal) {
				continue
			}
			n.Proposals[string(proposal.Id)] = unconfirmed
			if service.ViewManager.IsLeader() {
				ProposalMessageChan <- proposal
			} else {
				service.Net.SendToLeader(msgByte, service.ProposalMsg)
			}
		case msgByte := <-service.BlockChan:
			if service.ViewManager.IsOnChanging() {
				//TODO Add feedback mechanism which send msg to client
				continue
			}
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
			switch n.ValidateBlock(&msg) {
			case dataSync:
				continue
			case invalid:
				logger.Warningf("[Node.Run] block is invalid")
				continue
			}
			n.Block = msg.Block
			n.BlockPrepareMsg = map[string][]byte{}
			for _, p := range msg.AbandonedProposal {
				n.Proposals[string(p.Id)] = drop
			}
			for _, p := range msg.ProposalMessages {
				n.Proposals[string(p.Id)] = keep
			}
			n.Block = msg.Block
			blockConfirmMsg, err := messages.NewBlockConfirmMessage(id)
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
			var msg messages.BlockConfirmMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if !msg.Verify() {
				logger.Warningf("[Node.Run] msg.Verify failed")
				continue
			}
			n.BlockPrepareMsg[msg.From] = msg.Signature
			if service2.CertificateAuthorityX509.Check(len(n.BlockPrepareMsg)) {
				blockValidated := blockChain.NewBlockValidated(&n.Block, n.BlockPrepareMsg)
				if blockValidated == nil {
					logger.Warningf("[Node.Run] NewBlockValidated failed")
					continue
				}
				n.ExecuteBlock(blockValidated)
				if service.ViewManager.IsLeader() {
					BlockConfirmChan <- 1
				}
			}
		case msgByte := <-service.DataSyncChan:
			var msg messages.DataSyncMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if !service2.CertificateAuthorityX509.Exits(msg.From) {
				logger.Warningf("[Node.Run] DataSyncMessage.From is not exist")
				continue
			}
			if !msg.VerifySignature() {
				logger.Warningf("[Node.Run] DataSyncMessage.VerifySignature failed")
				continue
			}
			block, err := blockChain.BlockChain.GetBlockByHeight(msg.Height)
			if err != nil {
				logger.Warningf("[Node.Run] GetBlockByHeight error=%v", err)
				continue
			}
			respMsg, err := blockChain.NewDataSyncRespMessage(block)
			if err != nil {
				logger.Warningf("[Node.Run] NewDataSyncRespMessage error=%v", err)
				continue
			}
			jsonData, err := json.Marshal(respMsg)
			if err != nil {
				logger.Warningf("[Node.Run json.Marshal error=%v", err)
				continue
			}
			service.Net.SendTo(jsonData, service.DataSyncRespMsg, msg.From)
		case msgByte := <-service.DataSyncRespChan:
			var msg blockChain.DataSyncRespMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if !msg.Validate() {
				logger.Warningf("[Node.Run] DataSyncRespMessage.Validate failed")
				continue
			}
			n.ExecuteBlock(&msg.BlockValidated)
		case msgByte := <-service.ProposalConfirmChan:
			var msg messages.ProposalConfirm
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Node.Run] json.Unmarshal error=%v", err)
				continue
			}
			if state, ok := n.Proposals[string(msg.ProposalHash)]; !ok || state == drop {
				logger.Warningf("[Node.Run] I have never received this proposal")
				continue
			} else if state == unconfirmed {
				//TODO start view change
				lastBlock, err := blockChain.BlockChain.GetLatestBlock()
				if err != nil {
					logger.Warningf("[Node.Run] ProposalConfirm GetLatestBlock error=%v", err)
					continue
				}
				viewChangeMsg, err := blockChain.NewViewChangeMessage(lastBlock.Height,
					service.ViewManager.View, lastBlock.BlockHeader, lastBlock.Signatures)
				if err != nil {
					logger.Warningf("[Node.Run] ProposalConfirm NewViewChangeMessage error=%v", err)
					continue
				}
				jsonData, err := json.Marshal(viewChangeMsg)
				if err != nil {
					logger.Warningf("[Node.Run] ProposalConfirm json.Marshal error=%v", err)
					continue
				}
				service.Net.BroadCast(jsonData, service.ViewChangeMsg)
			} else {
				//This proposal is unready
			}
		}
	}
}

func (n *Node) handleProposal(proposal messages.ProposalMessage) bool {
	switch proposal.Type {
	case messages.Add:
		if !proposal.ValidateAdd() {
			logger.Warningf("[handleProposal] ValidateAdd failed")
			return false
		}
	case messages.Del:
		if !proposal.ValidateDel() {
			logger.Warningf("[handleProposal] ValidateDel failed")
			return false
		}
	case messages.Mod:
		if !proposal.ValidateMod() {
			logger.Warningf("[handleProposal] ValidateMod failed")
			return false
		}
	}
	return true
}

func (n *Node) ValidateBlock(msg *blockChain.BlockMessage) uint8 {
	lastBlock, err := blockChain.BlockChain.GetLatestBlock()
	if err != nil {
		logger.Warningf("[Node.Run] DataSync GetLatestBlock error=%v", err)
		return invalid
	}
	prevHash, err := lastBlock.Hash()
	if err != nil {
		logger.Warningf("[Node.Run] lastBlock.Hash error=%v", err)
		return invalid
	}
	if lastBlock.Height < msg.Block.Height-1 {
		service.StartDataSync(lastBlock.Height+1, msg.Block.Height-1)
		n.EnqueueBlock(*blockChain.NewBlockValidated(&msg.Block, map[string][]byte{}))
		return dataSync
	}
	if bytes.Compare(msg.Block.PrevBlock, prevHash) != 0 {
		logger.Warningf("[Node.Run] PrevBlock is invalid")
		return invalid
	}
	if !msg.VerifyBlock() {
		logger.Warningf("[ValidateBlock] VerifyBlock failed")
		return invalid
	}
	if !msg.VerifySignature() {
		logger.Warningf("[ValidateBlock] VerifySignature failed")
		return invalid
	}
	if !ValidateProposals(msg) {
		logger.Warningf("[ValidateBlock] ValidateProposals failed")
		return invalid
	}
	return ok
}

func (n *Node) EnqueueBlock(block blockChain.BlockValidated) {
	insert := false
	for i, b := range n.Blocks {
		if block.Height < b.Height {
			n.Blocks = append(n.Blocks[:i+1], n.Blocks[i:]...)
			n.Blocks[i] = block
			insert = true
			break
		} else if block.Height == b.Height {
			insert = true
			break
		}
	}
	if !insert {
		n.Blocks = append(n.Blocks, block)
	}
}

func (n *Node) ExecuteBlock(b *blockChain.BlockValidated) {
	lastBlock, err := blockChain.BlockChain.GetLatestBlock()
	if err != nil {
		logger.Warningf("[Node.Run] ExecuteBlock GetLatestBlock error=%v", err)
		return
	}
	if lastBlock.Height >= b.Height {
		return
	}
	n.EnqueueBlock(*b)
	h := lastBlock.Height + 1
	for _, bb := range n.Blocks {
		if bb.Height < h {
			n.Blocks = n.Blocks[1:]
		} else if bb.Height == h {
			err := blockChain.BlockChain.AddBlock(b)
			if err != nil {
				logger.Warningf("[Node.Run] ExecuteBlock AddBlock error=%v", err)
				return
			}
			n.SendReply(&b.Block)
			n.Blocks = n.Blocks[1:]
		} else {
			return
		}
		h++
	}
}

func (n *Node) SendReply(b *blockChain.Block) {
	for _, p := range b.ProposalMessages {
		msg, err := messages.NewProposalReplyMessage(p.Id)
		if err != nil {
			logger.Warningf("[SendReply] NewProposalReplyMessage error=%v", err)
			continue
		}
		jsonData, err := json.Marshal(msg)
		if err != nil {
			logger.Warningf("[SendReply] json.Marshal error=%v", err)
			continue
		}
		service.Net.SendTo(jsonData, service.ProposalReplyMsg, p.From)
	}
}
