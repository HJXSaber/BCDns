package service

import (
	"BCDns_0.1/network/service"
	"encoding/json"
)

var (
	NodeG NodeT
)

type NodeT struct {
	Proposals map[string]ProposalMessage
}

func (n *NodeT) Run() {
	for {
		select {
		case msgByte := <- service.ProposalChan:
			if service.Leader.IsLeader() {
				LeaderG.ProposalMessageChan <- msgByte
			} else {
				n.handleProposal(msgByte)
			}
		}
	}
}

func (n *NodeT) handleProposal(msgByte []byte) {
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
	n.Proposals[string(proposal.Id)] = proposal
	service.Net.SendToLeader(msgByte, service.ProposalMsgT)
}