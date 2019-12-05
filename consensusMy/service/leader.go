package service

import "encoding/json"

var (
	LeaderRole *LeaderT
)

type LeaderT struct {
	ProposalMessageChan chan []byte

}

func (l *LeaderT) Run() {
	for {
		select {
		case msgByte := <- l.ProposalMessageChan:
			var msg ProposalMessage
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				logger.Warningf("[Leader.Run] json.Unmarshal error=%v", err)
				continue
			}
			l.handleProposal(msg)
		}
	}
}

func (l *LeaderT) handleProposal(proposal ProposalMessage) {
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

}