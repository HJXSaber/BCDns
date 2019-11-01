package blockChain

import (
	"BCDns_0.1/certificateAuthority/service"
	"BCDns_0.1/messages"
	"BCDns_0.1/utils"
	"bytes"
	"encoding/gob"
	"reflect"
)

type BlockSlice []Block

func (bs BlockSlice) Exists(b Block) bool {

	//Traverse array in reverse order because if a block exists is more likely to be on top.
	l := len(bs)
	for i := l - 1; i >= 0; i-- {

		bb := bs[i]
		if reflect.DeepEqual(b.Signature, bb.Signature) {
			return true
		}
	}

	return false
}

func (bs BlockSlice) PreviousBlock() *Block {
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

type Block struct {
	*BlockHeader
	Signature []byte
	*messages.ProposalSlice
}

type BlockHeader struct {
	From       string
	PrevBlock  []byte
	MerkelRoot []byte
	Timestamp  uint32
	Nonce      uint32
}

func NewBlock(previousBlock []byte) Block {
	header := &BlockHeader{PrevBlock: previousBlock}
	return Block{header, nil, new(messages.ProposalSlice)}
}

func (b *Block) AddProposal(t *messages.ProposalMassage) {
	newSlice := b.ProposalSlice.AddProposal(*t)
	b.ProposalSlice = &newSlice
}

func (b *Block) Sign() []byte {
	s := service.CertificateAuthorityX509.Sign(b.Hash())
	return s
}

func (b *Block) VerifyBlock() bool {
	merkel := b.GenerateMerkelRoot()
	return reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot) &&
		service.CertificateAuthorityX509.VerifySignature(b.Signature, b.Hash(), b.From)
}

func (b *Block) Hash() []byte {
	headerHash, _ := b.BlockHeader.MarshalBinary()
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
		[]messages.ProposalMassage(*b.ProposalSlice)).([][]byte)
	if !ok {
		return nil
	}
	return merkell(ts)

}
func (b *Block) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(*b); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *Block) UnmarshalBinary(d []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	if err := dec.Decode(b); err != nil {
		return err
	}

	return nil
}

func (h *BlockHeader) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(*h); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *BlockHeader) UnmarshalBinary(d []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	err := dec.Decode(h)
	if err != nil {
		return err
	}

	return nil
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
