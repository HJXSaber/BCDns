package messages

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"testing"

	"github.com/rs/xid"
)

func TestUUID(t *testing.T) {
	id := xid.New()
	fmt.Println(id)
	data, err := json.Marshal(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(data)
	var id2 xid.ID
	err = json.Unmarshal(data, id2.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id2)
}

func TestProposalMassage_ValidateAdd(t *testing.T) {
	type fields struct {
		Body      ProposalBody
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
			p := &ProposalMassage{
				Body:      tt.fields.Body,
				Signature: tt.fields.Signature,
			}
			if got := p.ValidateAdd(); got != tt.want {
				t.Errorf("ProposalMassage.ValidateAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalMassage_DeepEqual(t *testing.T) {
	type fields struct {
		Body      ProposalBody
		Signature []byte
	}
	type args struct {
		pp ProposalMassage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ProposalMassage{
				Body:      tt.fields.Body,
				Signature: tt.fields.Signature,
			}
			if got := p.DeepEqual(tt.args.pp); got != tt.want {
				t.Errorf("ProposalMassage.DeepEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalMassage_Sign(t *testing.T) {
	type fields struct {
		Body      ProposalBody
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
			p := &ProposalMassage{
				Body:      tt.fields.Body,
				Signature: tt.fields.Signature,
			}
			if err := p.Sign(); (err != nil) != tt.wantErr {
				t.Errorf("ProposalMassage.Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProposalMassage_VerifySignature(t *testing.T) {
	type fields struct {
		Body      ProposalBody
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
			p := &ProposalMassage{
				Body:      tt.fields.Body,
				Signature: tt.fields.Signature,
			}
			if got := p.VerifySignature(); got != tt.want {
				t.Errorf("ProposalMassage.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPId_String(t *testing.T) {
	type fields struct {
		Name           string
		NodeId         int64
		SequenceNumber string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PId{
				Name:           tt.fields.Name,
				NodeId:         tt.fields.NodeId,
				SequenceNumber: tt.fields.SequenceNumber,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("PId.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want *ProposalMassage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProposal(t *testing.T) {
	type args struct {
		zoneName string
		t        OperationType
	}
	tests := []struct {
		name string
		args args
		want *ProposalMassage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProposal(tt.args.zoneName, tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProposal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalBody_Hash(t *testing.T) {
	type fields struct {
		Timestamp int64
		PId       PId
		Type      OperationType
		ZoneName  string
		Nonce     uint32
		Values    map[string]string
		Id        []byte
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
			p := &ProposalBody{
				Timestamp: tt.fields.Timestamp,
				PId:       tt.fields.PId,
				Type:      tt.fields.Type,
				ZoneName:  tt.fields.ZoneName,
				Nonce:     tt.fields.Nonce,
				Values:    tt.fields.Values,
				Id:        tt.fields.Id,
			}
			got, err := p.Hash()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProposalBody.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProposalBody.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalBody_GetPowHash(t *testing.T) {
	type fields struct {
		Timestamp int64
		PId       PId
		Type      OperationType
		ZoneName  string
		Nonce     uint32
		Values    map[string]string
		Id        []byte
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
			p := &ProposalBody{
				Timestamp: tt.fields.Timestamp,
				PId:       tt.fields.PId,
				Type:      tt.fields.Type,
				ZoneName:  tt.fields.ZoneName,
				Nonce:     tt.fields.Nonce,
				Values:    tt.fields.Values,
				Id:        tt.fields.Id,
			}
			if err := p.GetPowHash(); (err != nil) != tt.wantErr {
				t.Errorf("ProposalBody.GetPowHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProposalBody_ValidatePow(t *testing.T) {
	type fields struct {
		Timestamp int64
		PId       PId
		Type      OperationType
		ZoneName  string
		Nonce     uint32
		Values    map[string]string
		Id        []byte
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
			p := &ProposalBody{
				Timestamp: tt.fields.Timestamp,
				PId:       tt.fields.PId,
				Type:      tt.fields.Type,
				ZoneName:  tt.fields.ZoneName,
				Nonce:     tt.fields.Nonce,
				Values:    tt.fields.Values,
				Id:        tt.fields.Id,
			}
			if got := p.ValidatePow(); got != tt.want {
				t.Errorf("ProposalBody.ValidatePow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalPool_Len(t *testing.T) {
	type fields struct {
		Mutex         sync.Mutex
		ProposalSlice ProposalSlice
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &ProposalPool{
				Mutex:         tt.fields.Mutex,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if got := pool.Len(); got != tt.want {
				t.Errorf("ProposalPool.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalPool_Exits(t *testing.T) {
	type fields struct {
		Mutex         sync.Mutex
		ProposalSlice ProposalSlice
	}
	type args struct {
		pm ProposalMassage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &ProposalPool{
				Mutex:         tt.fields.Mutex,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			if got := pool.Exits(tt.args.pm); got != tt.want {
				t.Errorf("ProposalPool.Exits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalPool_Clear(t *testing.T) {
	type fields struct {
		Mutex         sync.Mutex
		ProposalSlice ProposalSlice
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &ProposalPool{
				Mutex:         tt.fields.Mutex,
				ProposalSlice: tt.fields.ProposalSlice,
			}
			pool.Clear()
		})
	}
}

func TestProposalSlice_FindByZoneName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		s    *ProposalSlice
		args args
		want *ProposalMassage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.FindByZoneName(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProposalSlice.FindByZoneName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProposalAuditResponse(t *testing.T) {
	type args struct {
		proposal ProposalMassage
	}
	tests := []struct {
		name string
		args args
		want *ProposalAuditResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProposalAuditResponse(tt.args.proposal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProposalAuditResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalAuditResponse_Verify(t *testing.T) {
	type fields struct {
		ProposalHash []byte
		Auditor      string
		Signature    []byte
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
			pr := &ProposalAuditResponse{
				ProposalHash: tt.fields.ProposalHash,
				Auditor:      tt.fields.Auditor,
				Signature:    tt.fields.Signature,
			}
			if got := pr.Verify(); got != tt.want {
				t.Errorf("ProposalAuditResponse.Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalAuditResponses_Check(t *testing.T) {
	tests := []struct {
		name string
		prs  *ProposalAuditResponses
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prs.Check(); got != tt.want {
				t.Errorf("ProposalAuditResponses.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAuditedProposal(t *testing.T) {
	type args struct {
		p         ProposalMassage
		responses ProposalAuditResponses
		termid    int64
	}
	tests := []struct {
		name    string
		args    args
		want    *AuditedProposal
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAuditedProposal(tt.args.p, tt.args.responses, tt.args.termid)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuditedProposal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAuditedProposal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuditedProposal_Hash(t *testing.T) {
	type fields struct {
		TermId     int64
		From       string
		Proposal   ProposalMassage
		Signatures map[string][]byte
		Signature  []byte
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
			m := AuditedProposal{
				TermId:     tt.fields.TermId,
				From:       tt.fields.From,
				Proposal:   tt.fields.Proposal,
				Signatures: tt.fields.Signatures,
				Signature:  tt.fields.Signature,
			}
			got, err := m.Hash()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuditedProposal.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuditedProposal.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuditedProposal_Sign(t *testing.T) {
	type fields struct {
		TermId     int64
		From       string
		Proposal   ProposalMassage
		Signatures map[string][]byte
		Signature  []byte
	}

	//p := NewProposal(".com", Add)
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"1",
			fields{
				TermId: 12,
				From:   "11",
				//Proposal:   *p,
				Signatures: map[string][]byte{},
			},
			[]byte{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AuditedProposal{
				TermId: tt.fields.TermId,
				From:   tt.fields.From,
				//Proposal:   tt.fields.Proposal,
				Signatures: tt.fields.Signatures,
				Signature:  tt.fields.Signature,
			}
			got, err := m.Sign()
			if (err != nil) != tt.wantErr {
				t.Errorf("AuditedProposal.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuditedProposal.Sign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuditedProposal_VerifySignature(t *testing.T) {
	type fields struct {
		TermId     int64
		From       string
		Proposal   ProposalMassage
		Signatures map[string][]byte
		Signature  []byte
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
			m := AuditedProposal{
				TermId:     tt.fields.TermId,
				From:       tt.fields.From,
				Proposal:   tt.fields.Proposal,
				Signatures: tt.fields.Signatures,
				Signature:  tt.fields.Signature,
			}
			if got := m.VerifySignature(); got != tt.want {
				t.Errorf("AuditedProposal.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuditedProposal_VerifySignatures(t *testing.T) {
	type fields struct {
		TermId     int64
		From       string
		Proposal   ProposalMassage
		Signatures map[string][]byte
		Signature  []byte
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
			m := AuditedProposal{
				TermId:     tt.fields.TermId,
				From:       tt.fields.From,
				Proposal:   tt.fields.Proposal,
				Signatures: tt.fields.Signatures,
				Signature:  tt.fields.Signature,
			}
			if got := m.VerifySignatures(); got != tt.want {
				t.Errorf("AuditedProposal.VerifySignatures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalResult_Hash(t *testing.T) {
	type fields struct {
		ProposalHash []byte
		From         string
		Signature    []byte
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
			p := ProposalResult{
				ProposalHash: tt.fields.ProposalHash,
				From:         tt.fields.From,
				Signature:    tt.fields.Signature,
			}
			got, err := p.Hash()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProposalResult.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProposalResult.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalResult_Sign(t *testing.T) {
	type fields struct {
		ProposalHash []byte
		From         string
		Signature    []byte
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
			p := ProposalResult{
				ProposalHash: tt.fields.ProposalHash,
				From:         tt.fields.From,
				Signature:    tt.fields.Signature,
			}
			got, err := p.Sign()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProposalResult.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProposalResult.Sign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProposalResult_VerifySignature(t *testing.T) {
	type fields struct {
		ProposalHash []byte
		From         string
		Signature    []byte
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
			p := ProposalResult{
				ProposalHash: tt.fields.ProposalHash,
				From:         tt.fields.From,
				Signature:    tt.fields.Signature,
			}
			if got := p.VerifySignature(); got != tt.want {
				t.Errorf("ProposalResult.VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProposalResult(t *testing.T) {
	type args struct {
		p ProposalMassage
	}
	tests := []struct {
		name    string
		args    args
		want    *ProposalResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProposalResult(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProposalResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProposalResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
