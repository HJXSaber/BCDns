package blockChain

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	service2 "BCDns_0.1/consensusMy/service"
	"BCDns_0.1/utils"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
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
	service2.ProposalMessages
}

type BlockValidated struct {
	Block
	Signatures map[string][]byte
}

type BlockHeader struct {
	PrevBlock  []byte
	MerkelRoot []byte
	Timestamp  int64
	Height     uint
}

func NewBlock(proposals service2.ProposalMessages, previousBlock []byte, height uint) *Block {
	header := BlockHeader{
		PrevBlock: previousBlock,
		Height:    height,
		Timestamp: time.Now().Unix(),
	}
	b := &Block{header, proposals}
	b.MerkelRoot = b.GenerateMerkelRoot()
	return b
}

func NewGenesisBlock() *Block {
	return NewBlock(service2.ProposalMessages{}, []byte{}, 0)
}

func (b *Block) VerifyBlock() bool {
	merkel := b.GenerateMerkelRoot()
	return reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot)
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

	ts, ok := Map(func(t service2.ProposalMessage) ([]byte, error) { return t.Id, nil },
		[]service2.ProposalMessage(b.ProposalMessages)).([][]byte)
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
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(h); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
	service2.Base
	Block
	AbandonedProposal service2.ProposalMessages
	Signature         []byte
}

//TODO
func NewBlockMessage(b *Block, abandonedP service2.ProposalMessages) (BlockMessage, error) {
	msg := BlockMessage{
		Base: service2.Base{
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
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(msg.Base); err != nil {
		return nil, err
	}
	if err := enc.Encode(AbandonedProposalPool); err != nil {
		return nil, err
	}
	if hash, err := msg.Block.Hash(); err != nil {
		return nil, err
	} else {
		_, err := buf.Write(hash)
		if err != nil {
			return nil, err
		}
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
