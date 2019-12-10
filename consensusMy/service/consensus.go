package service

import "errors"

var ConsensusCenter *Consensus

type Consensus struct {
	Proposer
	Node
	Leader
}

func NewConsensus(done chan uint) (*Consensus, error) {
	p := NewProposer()
	if p == nil {
		return nil, errors.New("[Consensus] NewProposer failed")
	}
	n := NewNode()
	if n == nil {
		return nil, errors.New("[Consensus] NewNode failed")
	}
	l := NewLeader()
	if l == nil {
		return nil, errors.New("[Consensus] NewLeader failed")
	}
	go p.Run(done)
	go n.Run(done)
	go l.Run(done)
	return &Consensus{
		*p,
		*n,
		*l,
	}, nil
}
