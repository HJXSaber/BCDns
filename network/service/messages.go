package service

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/utils"
	"bytes"
	"encoding/gob"
	"errors"
	"time"
)

type Message interface{}

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

type JoinMessage struct {
	utils.Base
	Cert []byte
	Signature []byte
}

func NewJoinMessage() (*JoinMessage, error) {
	_, cert := service.CertificateAuthorityX509.GetLocalCertificate()
	msg := &JoinMessage{
		Base:utils.Base{
			From:conf.BCDnsConfig.HostName,
			TimeStamp:time.Now().Unix(),
		},
		Cert:cert,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *JoinMessage) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(msg.Base); err != nil {
		return nil, err
	}
	if err := enc.Encode(msg.Cert); err != nil {
		return nil, err
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *JoinMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[JoinMessage] Generate signature failed")
}

func (msg *JoinMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}

type JoinReplyMessage struct {
	utils.Base
	View int
	Signatures map[string][]byte
	Signature []byte
}

func NewJoinReplyMessage(view int, signatures map[string][]byte) (*JoinReplyMessage, error) {
	msg := &JoinReplyMessage{
		Base:utils.Base{
			From:conf.BCDnsConfig.HostName,
			TimeStamp:time.Now().Unix(),
		},
		View:view,
		Signatures:signatures,
	}
}

func (msg *JoinReplyMessage) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(msg.Base); err != nil {
		return nil, err
	}
	if err := enc.Encode(msg.View); err != nil {
		return nil, err

	}
	if err := enc.Encode(msg.Signatures); err != nil {
		return nil, err
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *JoinReplyMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[JoinReplyMessage] Generate signature failed")
}

func (msg *JoinReplyMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}