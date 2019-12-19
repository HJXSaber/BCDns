package blockChain

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

const BlockMaxSize = 50

type BlockSlice []Block

//
//func (bs BlockSlice) Exists(b Block) bool {
//
//	//Traverse array in reverse order because if a block exists is more likely to be on top.
//	l := len(bs)
//	for i := l - 1; i >= 0; i-- {
//
//		bb := bs[i]
//		if reflect.DeepEqual(bb.Signature, bb.Signature) {
//			return true
//		}
//	}
//
//	return false
//}

func (bs BlockSlice) PreviousBlock() *Block {
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

type Block struct {
	BlockHeader
	messages.ProposalMessages
}

type BlockHeader struct {
	PrevBlock  []byte
	MerkelRoot []byte
	Timestamp  int64
	Height     uint
}

func NewBlock(proposals messages.ProposalMessages, previousBlock []byte, height uint, genesis bool) *Block {
	t := time.Now().Unix()
	if genesis {
		t = 0
	}
	header := BlockHeader{
		PrevBlock: previousBlock,
		Height:    height,
		Timestamp: t,
	}
	b := &Block{header, proposals}
	b.MerkelRoot = b.GenerateMerkelRoot()
	return b
}

func NewGenesisBlock() *Block {
	return NewBlock(messages.ProposalMessages{}, []byte{}, 0, true)
}

func (b *Block) VerifyBlock() bool {
	merkel := b.GenerateMerkelRoot()
	return bytes.Compare(merkel, b.MerkelRoot) == 0
}

func (b *Block) Hash() ([]byte, error) {
	headerHash, err := b.BlockHeader.MarshalBlockHeader()
	if err != nil {
		return nil, err
	}
	return utils.SHA256(headerHash), nil
}

func (b *Block) GenerateMerkelRoot() []byte {
	var merkell func(hashes [][]byte) []byte
	merkell = func(hashes [][]byte) []byte {

		l := len(hashes)
		if l == 0 {
			return nil
		}
		if l == 1 {
			return hashes[0]
		} else {

			if l%2 == 1 {
				return merkell([][]byte{merkell(hashes[:l-1]), hashes[l-1]})
			}

			bs := make([][]byte, l/2)
			for i, _ := range bs {
				j, k := i*2, (i*2)+1
				bs[i] = utils.SHA256(append(hashes[j], hashes[k]...))
			}
			return merkell(bs)
		}
	}

	ts, ok := Map(func(t messages.ProposalMessage) ([]byte, error) { return t.Id, nil },
		[]messages.ProposalMessage(b.ProposalMessages)).([][]byte)
	if !ok {
		return nil
	}
	return merkell(ts)

}

func (b *Block) MarshalBlock() ([]byte, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func UnmarshalBlock(d []byte) (*Block, error) {
	b := new(Block)
	err := json.Unmarshal(d, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (h *BlockHeader) MarshalBlockHeader() ([]byte, error) {
	jsonData, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func UnmarshalBlockHeader(d []byte) (*BlockHeader, error) {
	b := new(BlockHeader)
	err := json.Unmarshal(d, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Map(f interface{}, vs interface{}) interface{} {

	vf := reflect.ValueOf(f)
	vx := reflect.ValueOf(vs)

	l := vx.Len()

	tys := reflect.SliceOf(vf.Type().Out(0))
	vys := reflect.MakeSlice(tys, l, l)

	for i := 0; i < l; i++ {

		y := vf.Call([]reflect.Value{vx.Index(i)})
		vys.Index(i).Set(y[0])
	}

	return vys.Interface()
}

type BlockMessage struct {
	utils.Base
	Block
	AbandonedProposal messages.ProposalMessages
	Signature         []byte
}

//TODO
func NewBlockMessage(b *Block, abandonedP messages.ProposalMessages) (BlockMessage, error) {
	msg := BlockMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		Block:             *b,
		AbandonedProposal: abandonedP,
	}
	err := msg.Sign()
	if err != nil {
		return BlockMessage{}, err
	}
	return msg, nil
}

func (msg *BlockMessage) Hash() ([]byte, error) {
	buf := bytes.Buffer{}
	if hash, err := msg.Block.Hash(); err != nil {
		return nil, err
	} else {
		buf.Write(hash)
	}
	if jsonData, err := json.Marshal(msg.Base); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := json.Marshal(msg.AbandonedProposal); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *BlockMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[BlockMessage.Sign] generate signature failed")
}

func (msg *BlockMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}

type BlockValidated struct {
	Block
	Signatures map[string][]byte
}

func NewBlockValidated(b *Block, signatures map[string][]byte) *BlockValidated {
	msg := &BlockValidated{
		Block:      *b,
		Signatures: signatures,
	}
	return msg
}

func (b *BlockValidated) MarshalBlockValidated() ([]byte, error) {
	hash, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func UnMarshalBlockValidated(data []byte) (*BlockValidated, error) {
	b := new(BlockValidated)
	err := json.Unmarshal(data, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type DataSyncRespMessage struct {
	utils.Base
	BlockValidated
	Signature []byte
}

func NewDataSyncRespMessage(b *BlockValidated) (DataSyncRespMessage, error) {
	msg := DataSyncRespMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		BlockValidated: *b,
	}
	err := msg.Sign()
	if err != nil {
		return DataSyncRespMessage{}, err
	}
	return msg, nil
}

func (msg *DataSyncRespMessage) Hash() ([]byte, error) {
	buf := bytes.Buffer{}
	bHash, err := msg.BlockValidated.Hash()
	if err != nil {
		return nil, err
	}
	buf.Write(bHash)
	if jsonData, err := json.Marshal(msg.Base); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *DataSyncRespMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[DataSyncRespMessage] Generate signature failed")
}

func (msg *DataSyncRespMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}

func (msg *DataSyncRespMessage) Validate() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	if !service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From) {
		return false
	}
	count := len(msg.Signatures)
	for host, sig := range msg.Signatures {
		if !service.CertificateAuthorityX509.VerifySignature(sig, hash, host) {
			fmt.Println("???", host, sig)
			count--
			if !service.CertificateAuthorityX509.Check(count) {
				return false
			}
		}
	}
	return true
}

type ViewChangeMessage struct {
	utils.Base
	BlockHeader
	Signatures map[string][]byte
	View       int64
	Signature  []byte
}

func NewViewChangeMessage(n uint, view int64, header BlockHeader, signatures map[string][]byte) (
	*ViewChangeMessage, error) {
	msg := &ViewChangeMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		BlockHeader: header,
		Signatures:  signatures,
		View:        view,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *ViewChangeMessage) Hash() ([]byte, error) {
	buf := bytes.Buffer{}
	if jsonData, err := json.Marshal(msg.Base); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := msg.BlockHeader.MarshalBlockHeader(); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := json.Marshal(msg.Signatures); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := json.Marshal(msg.View); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *ViewChangeMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); sig != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[ViewChangeMessage] Generate signature failed")
}

func (msg *ViewChangeMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}

func (msg *ViewChangeMessage) VerifySignatures() bool {
	headerHash, err := msg.BlockHeader.MarshalBlockHeader()
	if err != nil {
		return false
	}
	hash := utils.SHA256(headerHash)
	for host, sig := range msg.Signatures {
		if !service.CertificateAuthorityX509.VerifySignature(sig, hash, host) {
			return false
		}
	}
	return true
}

type NewViewMessage struct {
	utils.Base
	View           int64
	ViewChangeMsgs map[string]ViewChangeMessage
	Signature      []byte
}

func NewNewViewMessage(msgs map[string]ViewChangeMessage, view int64) (*NewViewMessage, error) {
	msg := &NewViewMessage{
		Base: utils.Base{
			From:      conf.BCDnsConfig.HostName,
			TimeStamp: time.Now().Unix(),
		},
		View:           view,
		ViewChangeMsgs: msgs,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (msg *NewViewMessage) Hash() ([]byte, error) {
	buf := bytes.Buffer{}
	if jsonData, err := json.Marshal(msg.Base); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := json.Marshal(msg.View); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	if jsonData, err := json.Marshal(msg.ViewChangeMsgs); err != nil {
		return nil, err
	} else {
		buf.Write(jsonData)
	}
	return utils.SHA256(buf.Bytes()), nil
}

func (msg *NewViewMessage) Sign() error {
	hash, err := msg.Hash()
	if err != nil {
		return err
	}
	if sig := service.CertificateAuthorityX509.Sign(hash); err != nil {
		msg.Signature = sig
		return nil
	}
	return errors.New("[NewViewMessage] Generate signature failed")
}

func (msg *NewViewMessage) VerifySignature() bool {
	hash, err := msg.Hash()
	if err != nil {
		return false
	}
	return service.CertificateAuthorityX509.VerifySignature(msg.Signature, hash, msg.From)
}

func (msg *NewViewMessage) VerifyMsgs() bool {
	for _, msg := range msg.ViewChangeMsgs {
		if !msg.VerifySignature() {
			return false
		}
	}
	return true
}
