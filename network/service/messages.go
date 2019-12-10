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

type Message struct {
	MessageTypeT
	Payload []byte
}

func NewMessage(t MessageTypeT, payload []byte) Message {
	return Message{
		t,
		payload,
	}
}

type JoinMessage struct {
	utils.Base
	Cert      []byte
	NodeId    int64
	Signature []byte
}

func NewJoinMessage() (*JoinMessage, error) {
	_, cert := service.CertificateAuthorityX509.GetLocalCertificate()
	msg := &JoinMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		Cert:   cert,
		NodeId: service.CertificateAuthorityX509.NodeId,
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
	if err := enc.Encode(msg.NodeId); err != nil {
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
	View       int64
	Signatures map[string][]byte
	NodeId     int64
	Signature  []byte
}

func NewJoinReplyMessage(view int64, signatures map[string][]byte) (*JoinReplyMessage, error) {
	msg := &JoinReplyMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		View:       view,
		Signatures: signatures,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
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

type InitLeaderMessage struct {
	utils.Base
	NodeIds   []int64
	Signature []byte
}

func NewInitLeaderMessage(nodeIds []int64) (*InitLeaderMessage, error) {
	msg := &InitLeaderMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		NodeIds: nodeIds,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *InitLeaderMessage) Hash() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(msg.Base); err != nil {
		return nil, err
	}
	if err := enc.Encode(msg.NodeIds); err != nil {
		return nil, err
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *InitLeaderMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[InitLeaderMessage] Generate signature failed")
}

func (msg *InitLeaderMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}
