package blockChain

import (
	"BCDns_0.1/messages"
	"bytes"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"

var (
	BlockChain         = new(Blockchain)
	LeaderProposalPool = new(messages.ProposalPool)
	NodeProposalPool   = new(messages.ProposalPool)
)

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address, nodeID string) (*Blockchain, error) {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		return nil, errors.New("Blockchain already exists.")
	}

	var tip []byte

	genesis := NewGenesisBlock()

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Printf("[CreateBlockchain] error=%v\n", err)
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		bBytes, err := genesis.MarshalBinary()
		if err != nil {
			fmt.Printf("[CreateBlockchain] genesis.MarshalBinary error=%v\n", err)
			return err
		}
		err = b.Put(genesis.Hash(), bBytes)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), genesis.Hash())
		if err != nil {
			return err
		}
		tip = genesis.Hash()

		return nil
	})
	if err != nil {
		fmt.Printf("[CreateBlockchain] error=%v\n", err)
		return nil, err
	}

	bc := Blockchain{tip, db}

	return &bc, nil
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(nodeID string) (*Blockchain, error) {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("Blockchain already exists.")
		return nil, errors.New("Blockchain already exists.")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return nil, err
	}

	bc := Blockchain{tip, db}

	return &bc, nil
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash())

		if blockInDb != nil {
			return nil
		}

		blockData, err := block.MarshalBinary()
		if err != nil {
			return err
		}
		err = b.Put(block.Hash(), blockData)
		if err != nil {
			return err
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := new(Block)
		err = lastBlock.UnmarshalBinary(lastBlockData)
		if err != nil {
			return err
		}

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash())
			if err != nil {
				return err
			}
			bc.tip = block.Hash()
		}

		return nil
	})
	if err != nil {
		fmt.Printf("[AddBlock] error=%v\n", err)
		return err
	}
	return nil
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindProposal(ID []byte) (messages.ProposalMassage, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, p := range *block.ProposalSlice {
			if bytes.Compare(p.Body.Id, ID) == 0 {
				return p, nil
			}
		}

		if len(block.PrevBlock) == 0 {
			break
		}
	}

	return messages.ProposalMassage{}, errors.New("Transaction is not found")
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() (uint, error) {
	lastBlock := new(Block)

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		err := lastBlock.UnmarshalBinary(blockData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return lastBlock.Height, nil
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	block := new(Block)

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		err := block.UnmarshalBinary(blockData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return *block, err
	}

	return *block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash())

		if len(block.PrevBlock) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(proposals messages.ProposalSlice) (*Block, error) {
	var lastHash []byte
	var lastHeight uint

	for _, p := range proposals {
		// TODO: ignore transaction if it's not valid
		if !p.VerifySignature() {
			return nil, errors.New("[MineBlock] ERROR: Invalid Proposal")
		}
	}

	block := new(Block)
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		err := block.UnmarshalBinary(blockData)
		if err != nil {
			return err
		}

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		return nil, err
	}

	newBlock := NewBlock(proposals, lastHash, lastHeight+1)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData, err := newBlock.MarshalBinary()
		if err != nil {
			return err
		}
		err = b.Put(newBlock.Hash(), blockData)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), newBlock.Hash())
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newBlock, nil
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (bc *Blockchain) FindDomain(name string) (*messages.ProposalMassage, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		if p := block.ProposalSlice.FindByZoneName(name); p != nil {
			return p, nil
		}

		if len(block.PrevBlock) == 0 {
			break
		}
	}
	return nil, nil
}

func (bc *Blockchain) Get(key []byte) ([]byte, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		ps := ReverseSlice(*block.ProposalSlice)
		for _, p := range ps {
			if p.Body.ZoneName == string(key) {
				data, err := p.MarshalBinary()
				if err != nil {
					return nil, err
				}
				return data, nil
			}
		}

		if len(block.PrevBlock) == 0 {
			break

		}
	}
	return nil, leveldb.ErrNotFound
}

func (bc *Blockchain) Set(key, value []byte) error {
	return nil
}

func ReverseSlice(s messages.ProposalSlice) messages.ProposalSlice {
	ss := make(messages.ProposalSlice, len(s))
	for i, j := 0, len(s)-1; i <= j; i, j = i+1, j-1 {
		ss[i], ss[j] = s[j], s[i]
	}
	return ss
}
