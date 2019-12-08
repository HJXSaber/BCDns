package service

import (
	"BCDns_0.1/certificateAuthority/service"
	"bytes"
	"encoding/gob"
	"errors"
)

type Message interface{}

type MessageJoin struct {
	Cert      []byte
	Name      string
	Signature []byte
}

func (m MessageJoin) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(m.Cert); err != nil {
		return nil, err
	}
	if err := enc.Encode(m.Name); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m MessageJoin) Sign() ([]byte, error) {
	hash, err := m.Hash()
	if err != nil {
		return nil, err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		return sig, nil
	}
	return nil, errors.New("[MessageJoin] Generate signature failed")
}

func (m MessageJoin) VerifySignature() bool {
	hash, err := m.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(m.Signature, hash, m.Name)
}

type MessageProposal struct {
	Payload []byte
}

type MessageEndorsement struct {
	Payload []byte
}

type MessageCommit struct {
	Payload []byte
}

type MessageBlock struct {
	Payload []byte
}

type MessageReply struct {
	Payload []byte
}

type MessageViewChange struct {
	Payload []byte
}

type MessageRetrieveLeader struct {
	Payload []byte
}

type MessageProposalConfirm struct {
	Payload []byte
}

type MessageBlockConfirm struct {
	Payload []byte
}

type MessageDataSync struct {
	Payload []byte
}

type MessageDataSyncResp struct {
	Payload []byte
}

type MessageProposalReply struct {
	Payload []byte
}
