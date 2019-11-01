package service

import (
	"BCDns_0.1/blockChain"
	service2 "BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/network/service"
	"encoding/json"
	"fmt"
	"time"
)

type LeaderNode struct {
}

type LeaderNodeInterface interface {
	Run()
}

func (l LeaderNode) Run() {
	for true {
		select {
		case msgByte := <-service.CommitChan:
			var msg messages.AuditedProposal
			err := json.Unmarshal(msgByte, &msg)
			if err != nil {
				fmt.Printf("[LeaderNode] json.Unmarshal failed err=%v\n", err)
				continue
			}
			if !service.Leader.CheckTermId(msg.TermId) {
				fmt.Printf("[LeaderNode] termId is invalid\n")
				continue
			}
			if !service2.CertificateAuthorityX509.Exits(msg.From) {
				fmt.Printf("[LeaderNode] From node is unexist\n")
				continue
			}
			if !msg.VerifySignature() {
				fmt.Printf("[LeaderNode] Signature is illegal\n")
				continue
			}
			if !msg.VerifySignatures() {
				fmt.Printf("[LeaderNode] Signatures is illegal\n")
				continue
			}
			blockChain.ProposalPool.AddProposal(msg.Proposal)
		case <-time.After(10 * time.Second):
			blockBytes, err := blockChain.Bl.CurrentBlock.MarshalBinary()
			if err != nil {
				fmt.Printf("[LeaderNode] CurrentBlock marshal failed err=%v\n", err)
				continue
			}
			service.P2PNet.BroadcastMsg(blockBytes, service.Block)
		}
	}
}
