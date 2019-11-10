package blockChain

import (
	"BCDns_0.1/bcDns/conf"
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/utils"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

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

type BlockMessage struct {
	Block
	Signature []byte
}

type Block struct {
	BlockHeader
	messages.ProposalSlice
}

type BlockHeader struct {
	From       string
	PrevBlock  []byte
	MerkelRoot []byte
	Timestamp  int64
	Height     uint
}

func NewBlock(proposals messages.ProposalSlice, previousBlock []byte, height uint) *Block {
	header := BlockHeader{
		PrevBlock: previousBlock,
		Height:    height,
		From:      conf.BCDnsConfig.HostName,
		Timestamp: time.Now().Unix(),
	}
	b := &Block{header, proposals}
	b.MerkelRoot = b.GenerateMerkelRoot()
	return b
}

func NewGenesisBlock() *Block {
	return NewBlock(messages.ProposalSlice{}, []byte{}, 0)
}

func (b *Block) VerifyBlock() bool {
	merkel := b.GenerateMerkelRoot()
	return reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot)
}

func NewBlockMessage(b *Block) (*BlockMessage, error) {
	msg := &BlockMessage{
		Block: *b,
	}
	err := msg.Sign()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (b *BlockMessage) Sign() error {
	if sig := service.CertificateAuthorityX509.Sign(b.Hash()); sig != nil {
		b.Signature = sig
		return nil
	}
	return errors.New("Generate Signature failed")
}

func (b *BlockMessage) VerifySignature() bool {
	return service.CertificateAuthorityX509.VerifySignature(b.Signature, b.Hash(), b.From)
}

func (b *Block) Hash() []byte {
	headerHash, _ := b.BlockHeader.MarshalBlockHeader()
	return utils.SHA256(headerHash)
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

	ts, ok := Map(func(t messages.ProposalMassage) ([]byte, error) { return t.Body.Hash() },
		[]messages.ProposalMassage(b.ProposalSlice)).([][]byte)
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

		res := vf.Call([]reflect.Value{vx.Index(i)})
		y, err := res[0], res[1]
		if err.Interface() != nil {
			return nil
		}
		vys.Index(i).Set(y)
	}

	return vys.Interface()
}
