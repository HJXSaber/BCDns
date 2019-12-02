package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	service2 "BCDns_0.1/network/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"net"
	"sync"
	"time"
)

var (
	logger          *logging.Logger // package-level logger
)

type ProposerT struct {
	ReplyMutex sync.Mutex

	Proposals map[string]ProposalMessage
	Replys sync.Map
	Contexts sync.Map
	Conn            *net.UDPConn
	OrderChan       chan []byte
}

func NewProposer() *ProposerT {
	return &ProposerT{}
}

func init() {
	logger = logging.MustGetLogger("consensusMy/Proposer")
}

func (p *ProposerT) Run(done chan uint) {
	var (
		err error
	)
	defer close(done)
	go p.ReceiveOrder()
	for {
		select {
		case msgByte := <- p.OrderChan:
			var msg Order
			err = json.Unmarshal(msgByte, &msg)
			if err != nil {
				continue
			}
			p.handleOrder(msg)
		}
	}
}

type Order struct {
	OptType  messages.OperationType
	ZoneName string
	Values   map[string]string
}

func (p *ProposerT) ReceiveOrder() {
	var (
		data = make([]byte, 1024)
	)
	for true {
		len, err := p.Conn.Read(data)
		if err != nil {
			fmt.Printf("[Run] Proposer read order failed err=%v\n", err)
			continue
		}
		p.OrderChan <- data[:len]
	}
}

func (p *ProposerT) handleOrder(msg Order) {
	if proposal := messages.NewProposal(msg.ZoneName, msg.OptType, msg.Values); proposal != nil {
		proposalByte, err := json.Marshal(proposal)
		if err != nil {
			logger.Warningf("[handleOrder] json.Marshal error=%v", err)
			return
		}
		p.Replys.Store(string(proposal.Body.Id), map[string]uint8{})
		ctx := context.Background()
		go p.timer(ctx)
	}
}

func (p *ProposerT) timer(ctx context.Context, proposal ProposalMessage) {
	select {
	case <- time.After(conf.BCDnsConfig.ProposalTimeout):
		p.ReplyMutex.Lock()
		defer p.ReplyMutex.Unlock()
		repliesI, ok := p.Replys.Load(string(proposal.Id))
		if !ok {
			logger.Warningf("[Proposer.timer] Proposal is not exist")
			return
		}
		replies, ok := repliesI.(map[string]uint8)
		if !ok {
			logger.Warningf("[Proposer.timer] convert to map failed")
			return
		}
		if service.CertificateAuthorityX509.Check(len(replies)) {
			fmt.Printf("[Proposer.timer] Proposal=%v execute successfully", string(proposal.Id))
			p.Replys.Delete(string(proposal.Id))
		} else {
			confirmMsg := NewProposalConfirm(proposal.Id)
			if confirmMsg == nil {
				logger.Warningf("[Proposer.timer] NewProposalConfirm failed")
				return
			}
			confirmMsgByte, err := json.Marshal(confirmMsg)
			if err != nil {
				logger.Warningf("[Proposer.timer] json.Marshal error=%v", err)
				return
			}
			service2.Net.BroadCast(confirmMsgByte, service2.ProposalConfirmT)
		}
	case <- ctx.Done():
		fmt.Printf("[Proposer.timer] haha")
	}
}