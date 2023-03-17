package parlia

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const (
	wiggleTimeBeforeFork       = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
	fixedBackOffTimeBeforeFork = 200 * time.Millisecond
)

func (p *Parlia) delayForRamanujanFork(snap *Snapshot, header *types.Header) time.Duration {
	delay := time.Until(time.Unix(int64(header.Time), 0)) // nolint: gosimple
	if p.chainConfig.IsRamanujan(header.Number) {
		return delay
	}
	if header.Difficulty.Cmp(diffNoTurn) == 0 {
		// It's not our turn explicitly to sign, delay it a bit
		wiggle := time.Duration(len(snap.Validators)/2+1) * wiggleTimeBeforeFork
		delay += fixedBackOffTimeBeforeFork + time.Duration(rand.Int63n(int64(wiggle)))
	}
	return delay
}

func (p *Parlia) blockTimeForRamanujanFork(snap *Snapshot, header, parent *types.Header) uint64 {
	blockTime := parent.Time + p.config.Period
	if p.chainConfig.IsRamanujan(header.Number) {
		blockTime = blockTime + p.backOffTime(snap, header, p.val)
	}
	return blockTime
}

func (p *Parlia) blockTimeVerifyForRamanujanFork(snap *Snapshot, header, parent *types.Header) error {
	if p.chainConfig.IsRamanujan(header.Number) {
		backOffTime := p.backOffTime(snap, header, header.Coinbase)
		diffNoTurn := big.NewInt(1)
		if header.Difficulty.Cmp(diffNoTurn) == 0 && backOffTime == 0 {
			log.Info("verify, noTurn but backOffTime == 0", "addr", header.Coinbase,
				"number", header.Number, "GasLimit", header.GasLimit, "GasUsed", header.GasUsed)
		}
		if header.Time < parent.Time+p.config.Period+backOffTime {
			return consensus.ErrFutureBlock
		}
	}
	return nil
}
