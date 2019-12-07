package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/network/service"
	"encoding/json"
	"reflect"
)

const (
	keep uint8 = iota
	drop
)

const (
	ok uint8 = iota
	dataSync
	invalid
)

type Node struct {
	Proposals       map[string]uint8
	Blocks          []blockChain.Block
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
			n.Blocks = append(n.Blocks, msg.Block)
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
		case msgByte := <-service.DataSyncChan:
			var msg DataSyncMessage
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
			respMsg, err := NewDataSyncRespMessage(block)
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
		n.StartDataSync(lastBlock.Height+1, msg.Block.Height-1)
		n.EnqueueBlock(msg.Block)
		return dataSync
	}
	if reflect.DeepEqual(msg.Block.PrevBlock, prevHash) {
		logger.Warningf("[Node.Run] PrevBlock is invalid")
		return invalid
	}
	if !service2.CertificateAuthorityX509.Exits(msg.From) {
		logger.Warningf("[ValidateBlock] msg.From is not exist")
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

func (n *Node) StartDataSync(lastH, h uint) {
	for i := lastH; i <= h; i++ {
		syncMsg, err := NewDataSyncMessage(i)
		if err != nil {
			logger.Warningf("[DataSync] NewDataSyncMessage error=%v", err)
			continue
		}
		jsonData, err := json.Marshal(syncMsg)
		if err != nil {
			logger.Warningf("[DataSync] json.Marshal error=%v", err)
			continue
		}
		service.Net.BroadCast(jsonData, service.DataSyncMsg)
	}
}

func (n *Node) EnqueueBlock(block blockChain.Block) {
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
