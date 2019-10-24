package service

import (
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Proposer struct {
	TimeOut        time.Duration
	Conn           *net.UDPConn
	Proposals      map[string]*messages.ProposalMassage
	AuditResponses map[string]messages.ProposalAuditResponses
}

type ProposerInterface interface {
	Run()
}

type Order struct {
	OptType  messages.OperationType
	ZoneName string
}

func (p *Proposer) Run() {
	var (
		data = make([]byte, 1024)
	)
	for true {
		_, err := p.Conn.Read(data)
		if err != nil {
			fmt.Printf("[Run] Proposer read order failed err=%v\n", err)
			continue
		}
		go p.handleOrder(data)
	}
}

func (p *Proposer) handleOrder(data []byte) {
	var order Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		fmt.Printf("[handleOrder] json.Unmarshal failed err=%v\n", err)
		return
	}
	if proposal := messages.NewProposal(order.ZoneName, order.OptType); proposal != nil {
		proposalByte, err := json.Marshal(*proposal)
		if err != nil {
			fmt.Printf("[handleOrder] json.Marshal failed err=%v\n", err)
			return
		}
		p.Proposals[string(proposal.Body.HashCode)] = proposal
		service.P2PNet.BroadcastMsg(proposalByte, service.Proposal)
		done := make(chan int)
		go func() {
			select {
			case <-time.After(p.TimeOut):

			case <-done:
				auditedResponse := messages.NewAuditedProposal(*proposal, p.AuditResponses[proposal.Body.ZoneName])

			}
		}()
		if _, ok := p.AuditResponses[proposal.Body.ZoneName]; ok {
			p.AuditResponses[proposal.Body.ZoneName] = messages.ProposalAuditResponses{}
		}
		for true {
			select {
			case AuditResultByte := <-service.AuditResponseChan:
				var AuditResult messages.ProposalAuditResponse
				err = json.Unmarshal(AuditResultByte, &AuditResult)
				if err != nil {
					fmt.Printf("[handleOrder] json.Unmarshal failed err=%v\n", err)
					continue
				}
				p.AuditResponses[proposal.Body.ZoneName][AuditResult.Auditor] = AuditResult
				if p.AuditResponses[proposal.Body.ZoneName].Check() {
					close(done)
				}
			}
		}
	} else {
		fmt.Printf("[handleOrder] Generate proposal failed\n")
	}
}

func NewProposer(timeOut time.Duration) *Proposer {
	addr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8888,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("[NewProposer] Listen failed err=%v\n", err)
		return nil
	}
	return &Proposer{
		TimeOut: timeOut,
		Conn:    conn,
	}
}
