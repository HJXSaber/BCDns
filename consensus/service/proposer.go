package service

import (
	"BCDns_0.1/certificateAuthority/model"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"github.com/fanliao/go-concurrentMap"
	"net"
	"time"
)

type Proposer struct {
	TimeOut time.Duration
	Conn    *net.UDPConn
	//Proposals      map[string]*messages.ProposalMassage
	Proposals *concurrent.ConcurrentMap
	//AuditResponseChan map[string]chan messages.ProposalAuditResponses
	AuditResponseChan *concurrent.ConcurrentMap
	AuditResponses    map[string]messages.ProposalAuditResponses
	ProposalResults   map[string]map[string]uint8
	OrderChan chan []byte
	Timers map[string]*time.Timer
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
		select {
		case msgByte := <- p.OrderChan:
			p.handleOrder(msgByte)
		case msgByte := <-
		}
	}
}

func (p *Proposer) ReceiveOrder() {
	var (
		data = make([]byte, 1024)
	)
	for true {
		_, err := p.Conn.Read(data)
		if err != nil {
			fmt.Printf("[Run] Proposer read order failed err=%v\n", err)
			continue
		}
		p.OrderChan <- data
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
		_, err = p.Proposals.Put(string(proposal.Body.HashCode), proposal)
		if err != nil {
			fmt.Printf("[handleOrder] concurrentMap error=%v\n", err)
			return
		}
		service.P2PNet.BroadcastMsg(proposalByte, service.Proposal)
		p.Timers[string(proposal.Body.HashCode)] = time.AfterFunc(p.TimeOut, func() {

		})




		done := make(chan messages.ProposalAuditResponses)
		_, err = p.AuditResponseChan.Put(string(proposal.Body.HashCode), done)
		if err != nil {
			fmt.Printf("[handleOrder] ConcurrentMap error=%v\n", err)
			return
		}
		p.AuditResponses[proposal.Body.ZoneName] = messages.ProposalAuditResponses{}
		select {
		case <-time.After(p.TimeOut):
			goto clean
		case responses := <-done:
			auditedResponse, err := messages.NewAuditedProposal(*proposal, responses, service.Leader.TermId)
			if err != nil {
				fmt.Printf("[handleOrder] NewAuditedProposal err=%v\n", err)
				return
			}
			p.Commit(auditedResponse)
			close(done)
			_, err = p.AuditResponseChan.Remove(string(proposal.Body.HashCode))
			if err != nil {
				fmt.Printf("[handleOrder] ConcurrentMap error=%v\n", err)
				return
			}
		}
		select {
		case <-time.After(p.TimeOut):
			goto clean
		case msgByte := <-service.ProposalResultChan:
			var msg messages.ProposalResult

		}
	clean:
	} else {
		fmt.Printf("[handleOrder] Generate proposal failed\n")
	}
}

func (p *Proposer) ProcessAuditResponse() {
	for true {
		select {
		case AuditResultByte := <-service.AuditResponseChan:
			var AuditResult messages.ProposalAuditResponse
			err := json.Unmarshal(AuditResultByte, &AuditResult)
			if err != nil {
				fmt.Printf("[ProcessAuditResponse] json.Unmarshal failed err=%v\n", err)
				continue
			}
			proposalI, err := p.Proposals.Get(string(AuditResult.ProposalHash))
			if err != nil {
				fmt.Printf("[ProcessAuditResponse] ConcurrentMap error=%v\n", err)
				continue
			}
			proposal, ok := proposalI.(messages.ProposalMassage)
			if !ok {
				continue
			}
			if _, ok := p.AuditResponses[proposal.Body.ZoneName]; ok {
				p.AuditResponses[proposal.Body.ZoneName][AuditResult.Auditor] = AuditResult
				if p.AuditResponses[proposal.Body.ZoneName].Check() {
					auditResponseChanI, err := p.AuditResponseChan.Get(string(AuditResult.ProposalHash))
					if err != nil {
						fmt.Printf("[ProcessAuditResponse] ConcurrentMap error=%v\n", err)
						continue
					}
					auditResponseChan, ok := auditResponseChanI.(chan messages.ProposalAuditResponses)
					if !ok {
						continue
					}
					auditResponseChan <- p.AuditResponses[proposal.Body.ZoneName]
					delete(p.AuditResponses, proposal.Body.ZoneName)
				}
			}
		}
	}
}

func (p *Proposer) ProcessProposalResult() {
	var msg messages.ProposalResult
	for true {
		msgByte := <-service.ProposalResultChan
		err := json.Unmarshal(msgByte, &msg)
		if err != nil {
			fmt.Printf("[ProcessProposalResult] json.Unmarshal failed err=%v\n", err)
			continue
		}
		if _, ok := p.ProposalResults[string(msg.ProposalHash)]; ok {
			p.ProposalResults[string((msg.ProposalHash))][msg.From] = 0
			if service2.CertificateAuthorityX509.Check(len(p.ProposalResults[string(msg.ProposalHash)])) {

			}
		} else {

		}
	}
}

func (p *Proposer) Commit(data *messages.AuditedProposal) {
	jsonData, err := json.Marshal(*data)
	if err != nil {
		fmt.Printf("[Commit] json.Marshal failed err=%v\n", err)
		return
	}
	service.P2PNet.SendToLeader(jsonData, service.Commit)
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
		TimeOut:   timeOut,
		Conn:      conn,
		Proposals: concurrent.NewConcurrentMap(),
		//AuditResponseChan: make(map[string]chan messages.ProposalAuditResponses),
		AuditResponseChan: concurrent.NewConcurrentMap(),
		AuditResponses:    make(map[string]messages.ProposalAuditResponses),
		OrderChan:make(chan []byte, 1024),
	}
}
