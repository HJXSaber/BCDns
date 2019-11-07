package blockChain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"reflect"
	"testing"

	"BCDns_0.1/messages"
)

func TestBlockSlice_PreviousBlock(t *testing.T) {
	tests := []struct {
		name string
		bs   BlockSlice
		want *Block
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bs.PreviousBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockSlice.PreviousBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlock(t *testing.T) {
	type args struct {
		proposals     messages.ProposalSlice
		previousBlock []byte
		height        uint
	}
	tests := []struct {
		name string
		args args
		want *Block
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlock(tt.args.proposals, tt.args.previousBlock, tt.args.height); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGenesisBlock(t *testing.T) {
	tests := []struct {
		name string
		want *Block
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGenesisBlock(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGenesisBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_VerifyBlock(t *testing.T) {
	type fields struct {
		BlockHeader   BlockHeader
		ProposalSlice messages.ProposalSlice
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				BlockHeader:   tt.fields.BlockHeader,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if got := b.VerifyBlock(); got != tt.want {
				t.Errorf("Block.VerifyBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlockMessage(t *testing.T) {
	type args struct {
		b *Block
	}
	tests := []struct {
		name    string
		args    args
		want    *BlockMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBlockMessage(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlockMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockMessage_Sign(t *testing.T) {
	type fields struct {
		Block     Block
		Signature []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockMessage{
				Block:     tt.fields.Block,
				Signature: tt.fields.Signature,
			}
			if err := b.Sign(); (err != nil) != tt.wantErr {
				t.Errorf("BlockMessage.Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockMessage_VerifySignature(t *testing.T) {
	type fields struct {
		Block     Block
		Signature []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BlockMessage{
				Block:     tt.fields.Block,
				Signature: tt.fields.Signature,
			}
			if got := b.VerifySignature(); got != tt.want {
				t.Errorf("BlockMessage.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_Hash(t *testing.T) {
	type fields struct {
		BlockHeader   BlockHeader
		ProposalSlice messages.ProposalSlice
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				BlockHeader:   tt.fields.BlockHeader,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if got := b.Hash(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_GenerateMerkelRoot(t *testing.T) {
	type fields struct {
		BlockHeader   BlockHeader
		ProposalSlice messages.ProposalSlice
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				BlockHeader:   tt.fields.BlockHeader,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if got := b.GenerateMerkelRoot(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block.GenerateMerkelRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

type Block1 struct {
	H BlockHeader
}

func (b *Block1) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func TestBlock_MarshalBinary(t *testing.T) {
	type fields struct {
		BlockHeader BlockHeader
	}
	b := NewGenesisBlock()
	fmt.Println(b)
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", fields{
			BlockHeader: b.BlockHeader,
		}, []byte{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block1{
				H: tt.fields.BlockHeader,
			}
			got := b.Serialize()
			fmt.Println("Bytes:", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block.MarshalBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlock_UnmarshalBinary(t *testing.T) {
	type fields struct {
		BlockHeader   BlockHeader
		ProposalSlice messages.ProposalSlice
	}
	type args struct {
		d []byte
	}
	b := NewGenesisBlock()
	data, err := b.Marshal()
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"1", fields{
			BlockHeader:   b.BlockHeader,
			ProposalSlice: []messages.ProposalMassage{},
		}, args{
			data,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Block{
				BlockHeader:   tt.fields.BlockHeader,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if err := b.Unmarshal(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("Block.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlockHeader_MarshalBinary(t *testing.T) {
	type fields struct {
		ProposalSlice messages.ProposalSlice
		From          string
		PrevBlock     []byte
		MerkelRoot    []byte
		Timestamp     uint32
		Height        uint
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &BlockHeader{
				From:       tt.fields.From,
				PrevBlock:  tt.fields.PrevBlock,
				MerkelRoot: tt.fields.MerkelRoot,
				Timestamp:  tt.fields.Timestamp,
				Height:     tt.fields.Height,
			}
			got, err := h.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockHeader.MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockHeader.MarshalBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockHeader_UnmarshalBinary(t *testing.T) {
	type fields struct {
		ProposalSlice messages.ProposalSlice
		From          string
		PrevBlock     []byte
		MerkelRoot    []byte
		Timestamp     uint32
		Height        uint
	}
	type args struct {
		d []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &BlockHeader{
				From:       tt.fields.From,
				PrevBlock:  tt.fields.PrevBlock,
				MerkelRoot: tt.fields.MerkelRoot,
				Timestamp:  tt.fields.Timestamp,
				Height:     tt.fields.Height,
			}
			if err := h.Unmarshal(tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("BlockHeader.UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMap(t *testing.T) {
	type args struct {
		f  interface{}
		vs interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.f, tt.args.vs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}
