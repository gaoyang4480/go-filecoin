package blocksub

import (
	"context"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"

	"github.com/filecoin-project/go-filecoin/internal/pkg/block"
	"github.com/filecoin-project/go-filecoin/internal/pkg/consensus"
	"github.com/filecoin-project/go-filecoin/internal/pkg/metrics"
)

var blockTopicLogger = log.Logger("net/block_validator")
var mDecodeBlkFail = metrics.NewInt64Counter("net/pubsub_block_decode_failure", "Number of blocks that fail to decode seen on block pubsub channel")
var mInvalidBlk = metrics.NewInt64Counter("net/pubsub_invalid_block", "Number of blocks that fail syntax validation seen on block pubsub channel")

// BlockTopicValidator may be registered on go-libp2p-pubsub to validate blocksub messages.
type BlockTopicValidator struct {
	validator pubsub.Validator
	opts      []pubsub.ValidatorOpt
}

// NewBlockTopicValidator retruns a BlockTopicValidator using `bv` for message validation
func NewBlockTopicValidator(bv consensus.BlockSyntaxValidator, opts ...pubsub.ValidatorOpt) *BlockTopicValidator {
	return &BlockTopicValidator{
		opts: opts,
		validator: func(ctx context.Context, p peer.ID, msg *pubsub.Message) bool {
			blk, err := block.DecodeBlock(msg.GetData())
			if err != nil {
				blockTopicLogger.Debugf("block from peer: %s failed to decode: %s", p.String(), err.Error())
				mDecodeBlkFail.Inc(ctx, 1)
				return false
			}
			if err := bv.ValidateSyntax(ctx, blk); err != nil {
				blockTopicLogger.Debugf("block: %s from peer: %s failed to validate: %s", blk.Cid().String(), p.String(), err.Error())
				mInvalidBlk.Inc(ctx, 1)
				return false
			}
			return true
		},
	}
}

func (btv *BlockTopicValidator) Topic(network string) string {
	return Topic(network)
}

func (btv *BlockTopicValidator) Validator() pubsub.Validator {
	return btv.validator
}

func (btv *BlockTopicValidator) Opts() []pubsub.ValidatorOpt {
	return btv.opts
}