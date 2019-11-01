package blockChain

import (
	"fmt"
	"github.com/izqui/helpers"
	"time"
)

type BlocksQueue chan Block

var (
	Bl *Blockchain
)

type Blockchain struct {
	CurrentBlock Block
	BlockSlice

	BlocksQueue
}

func SetupBlockchan() *Blockchain {

	bl := new(Blockchain)
	bl.BlocksQueue = make(BlocksQueue)

	//Read blockchain from file and stuff...

	bl.CurrentBlock = bl.CreateNewBlock()

	return bl
}

func (bl *Blockchain) CreateNewBlock() Block {

	prevBlock := bl.BlockSlice.PreviousBlock()
	prevBlockHash := []byte{}
	if prevBlock != nil {
		prevBlockHash = prevBlock.Hash()
	}

	b := NewBlock(prevBlockHash)
	return b
}

func (bl *Blockchain) AddBlock(b Block) {
	bl.BlockSlice = append(bl.BlockSlice, b)
}

func (bl *Blockchain) GenerateBlocks() chan Block {


	interrupt := make(chan Block)

	go func() {

		block := <-interrupt
	loop:
		fmt.Println("Starting Proof of Work...")
		block.BlockHeader.MerkelRoot = block.GenerateMerkelRoot()
		block.BlockHeader.Nonce = 0
		block.BlockHeader.Timestamp = uint32(time.Now().Unix())

		for true {

			sleepTime := time.Nanosecond
			if block.TransactionSlice.Len() > 0 {

				if CheckProofOfWork(BLOCK_POW, block.Hash()) {

					block.Signature = block.Sign(Core.Keypair)
					bl.BlocksQueue <- block
					sleepTime = time.Hour * 24
					fmt.Println("Found Block!")

				} else {

					block.BlockHeader.Nonce += 1
				}

			} else {
				sleepTime = time.Hour * 24
				fmt.Println("No trans sleep")
			}

			select {
			case block = <-interrupt:
				goto loop
			case <-helpers.Timeout(sleepTime):
				continue
			}
		}
	}()

	return interrupt
}
