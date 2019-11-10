package service

import (
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"github.com/fanliao/go-concurrentMap"
	"net"
	"sync"
	"time"
)

var (
	Proposer *ProposerT
)

type ProposerT struct {
	TimeOut         time.Duration
	Conn            *net.UDPConn
	Proposals       map[string]*messages.ProposalMassage
	AuditResponses  *concurrent.ConcurrentMap
	ProposalResults *concurrent.ConcurrentMap
	OrderChan       chan []byte
	Timers          map[string]*time.Timer

	Mutex sync.Mutex
}

type ProposerInterface interface {
	Run()
}

type Order struct {
	OptType  messages.OperationType
	ZoneName string
	Values   map[string]string
}

func (p *ProposerT) Run(done chan uint) {
	defer close(done)
	go p.ReceiveOrder()
	for true {
		select {
		case msgByte := <-p.OrderChan:
			p.handleOrder(msgByte)
		case msgByte := <-service.AuditResponseChan:
			var msg messages.ProposalAuditResponse
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				fmt.Printf("[Proposer.Run] json.Unmarshal failed err=%v\n", err)
				continue
			}
			proposal, ok := p.Proposals[string(msg.ProposalHash)]
			if !ok {
				fmt.Printf("[Proposer.Run] proposal unexist %v", msg)
			}
			responsesI, err := p.AuditResponses.Get(string(msg.ProposalHash))
			if err != nil {
				fmt.Printf("[Proposer.Run] ConcurrentMap error=%v\n", err)
				continue
			}
			if responses, ok := responsesI.(messages.ProposalAuditResponses); ok {
				responses[msg.Auditor] = msg
				if service2.CertificateAuthorityX509.Check(len(responses)) {
					p.Mutex.Lock()
					p.Timers[string(msg.ProposalHash)].Stop()
					delete(p.Timers, string(string(msg.ProposalHash)))
					p.Mutex.Unlock()
					auditedResponse, err := messages.NewAuditedProposal(*proposal, responses, service.Leader.TermId)
					if err != nil {
						fmt.Printf("[Proposer.Run] NewAuditedProposal err=%v\n", err)
						continue
					}
					p.Commit(auditedResponse)
					_, err = p.ProposalResults.Put(string(msg.ProposalHash), map[string]uint8{})
					if err != nil {
						fmt.Printf("[Proposer.Run] ConcurrentMap error=%v\n", err)
						continue
					}
					p.Timers[string(msg.ProposalHash)] = time.AfterFunc(p.TimeOut, func() {
						p.Mutex.Lock()
						defer p.Mutex.Unlock()
						resultsI, err := p.ProposalResults.Get(string(msg.ProposalHash))
						if err != nil {
							fmt.Printf("[Proposer.Run] AuditResponseChan concurrentMap error=%v\n", err)
							return
						}
						if results, ok := resultsI.(map[string]uint8); ok {
							if service2.CertificateAuthorityX509.Check(len(results)) {
								fmt.Printf("[Proposer.Run] Proposal execute successfully %v\n", proposal)
							} else {
								//TODO: Proposal execute failed
								fmt.Printf("[Proposer.Run] Proposal execute failed %v\n", proposal)
							}
						}
						_, _ = p.ProposalResults.Remove(string(msg.ProposalHash))
					})
				}
			}
		case msgByte := <-service.ProposalResultChan:
			var msg messages.ProposalResult
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				fmt.Printf("[ProcessProposalResult] json.Unmarshal failed err=%v\n", err)
				continue
			}
			proposal, ok := p.Proposals[string(msg.ProposalHash)]
			if !ok {
				fmt.Printf("[Proposer.Run] proposal unexist %v", msg)
			}
			resultsI, err := p.ProposalResults.Get(string(msg.ProposalHash))
			if err != nil {
				fmt.Printf("[Proposer.Run] AuditResponseChan concurrentMap error=%v\n", err)
				return
			}
			if results, ok := resultsI.(map[string]uint8); ok {
				results[msg.From] = 0
				if service2.CertificateAuthorityX509.Check(len(results)) {
					p.Mutex.Lock()
					p.Timers[string(msg.ProposalHash)].Stop()
					delete(p.Timers, string(msg.ProposalHash))
					p.Mutex.Unlock()
					fmt.Printf("[Proposer.Run] Proposal execute successfully %v\n", proposal)
				}
			}
		}
	}
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

func (p *ProposerT) handleOrder(data []byte) {
	var order Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		fmt.Printf("[handleOrder] json.Unmarshal failed err=%v\n", err)
		return
	}
	if proposal := messages.NewProposal(order.ZoneName, order.OptType, order.Values); proposal != nil {
		proposalByte, err := json.Marshal(*proposal)
		if err != nil {
			fmt.Printf("[handleOrder] json.Marshal failed err=%v\n", err)
			return
		}
		proposalHash, err := proposal.Body.Hash()
		if err != nil {
			fmt.Printf("[NewProposalAuditResponse] generate failed err=%v\n", err)
			return
		}
		p.Proposals[string(proposalHash)] = proposal
		service.P2PNet.BroadcastMsg(proposalByte, service.Proposal)
		_, err = p.AuditResponses.Put(string(proposalHash), messages.ProposalAuditResponses{})
		if err != nil {
			fmt.Printf("[handleOrder] ConcurrentMap error=%v\n", err)
			return
		}
		p.Timers[string(string(proposalHash))] = time.AfterFunc(p.TimeOut, func() {
			p.Mutex.Lock()
			defer p.Mutex.Unlock()
			responsesI, err := p.AuditResponses.Get(string(proposalHash))
			if err != nil {
				fmt.Printf("[handleOrder] ConcurrentMap error=%v\n", err)
				return
			}
			if responses, ok := responsesI.(messages.ProposalAuditResponses); ok {
				if service2.CertificateAuthorityX509.Check(len(responses)) {
					auditedResponse, err := messages.NewAuditedProposal(*proposal, responses, service.Leader.TermId)
					if err != nil {
						fmt.Printf("[handleOrder] NewAuditedProposal err=%v\n", err)
						return
					}
					p.Commit(auditedResponse)
				} else {
					//TODO: Collect endorsement failed
					fmt.Printf("[Proposer.Run] Collect endorsement failed %v\n", proposal)
				}
			}
			_, _ = p.AuditResponses.Remove(string(proposalHash))
		})
	} else {
		fmt.Printf("[handleOrder] Generate proposal failed\n")
	}
}

func (p *ProposerT) Commit(data *messages.AuditedProposal) {
	jsonData, err := json.Marshal(*data)
	if err != nil {
		fmt.Printf("[Commit] json.Marshal failed err=%v\n", err)
		return
	}
	service.P2PNet.SendToLeader(jsonData, service.Commit)
}

func NewProposer(timeOut time.Duration) *ProposerT {
	addr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8888,
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("[NewProposer] Listen failed err=%v\n", err)
		return nil
	}
	return &ProposerT{
		TimeOut:         timeOut,
		Conn:            conn,
		Proposals:       map[string]*messages.ProposalMassage{},
		AuditResponses:  concurrent.NewConcurrentMap(),
		ProposalResults: concurrent.NewConcurrentMap(),
		OrderChan:       make(chan []byte, 1024),
		Timers:          map[string]*time.Timer{},
	}
}
