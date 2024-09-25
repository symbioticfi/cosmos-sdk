package abci

import (
	"cosmossdk.io/log"
	keeper2 "cosmossdk.io/x/symStaking/keeper"
	"cosmossdk.io/x/symStaking/types"
	"encoding/json"
	"errors"
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

const (
	SYMBIOTIC_SYNC_PERIOD = 10
	SLEEP_ON_RETRY        = 200
	RETRIES               = 5
)

type ProposalHandler struct {
	logger        log.Logger
	keeper        *keeper2.Keeper
	prevBlockTime uint64
}

func NewProposalHandler(logger log.Logger, keeper *keeper2.Keeper) *ProposalHandler {
	return &ProposalHandler{
		logger: logger,
		keeper: keeper,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
		var err error
		var resp *abci.PrepareProposalResponse

		for i := 0; i < RETRIES; i++ {
			resp, err = h.internalPrepareProposal(ctx, req)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * SLEEP_ON_RETRY)
		}

		if err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		return resp, nil
	}
}

func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(context sdk.Context, req *abci.FinalizeBlockRequest) error {
		var err error

		for i := 0; i < RETRIES; i++ {
			err = h.internalPreBlocker(context, req)
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * SLEEP_ON_RETRY)
		}

		if err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		return nil
	}
}

func (h *ProposalHandler) internalPrepareProposal(ctx sdk.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	proposalTxs := req.Txs

	if req.Height%SYMBIOTIC_SYNC_PERIOD != 0 {
		return &abci.PrepareProposalResponse{
			Txs: proposalTxs,
		}, nil
	}

	blockHash, err := h.keeper.GetFinalizedBlockHash(ctx)
	if err != nil {
		return nil, err
	}

	// NOTE: We use stdlib JSON encoding, but an application may choose to use
	// a performant mechanism. This is for demo purposes only.
	bz, err := json.Marshal(blockHash)
	if err != nil {
		return nil, errors.New("failed to encode injected vote extension tx")
	}

	// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
	// and store the canonical stake-weighted average prices.
	proposalTxs = append([][]byte{bz}, proposalTxs...)

	return &abci.PrepareProposalResponse{
		Txs: proposalTxs,
	}, nil
}

func (h *ProposalHandler) internalPreBlocker(context sdk.Context, req *abci.FinalizeBlockRequest) error {
	if req.Height%SYMBIOTIC_SYNC_PERIOD != 0 || len(req.Txs) == 0 {
		return nil
	}

	var blockHash string
	if err := json.Unmarshal(req.Txs[0], &blockHash); err != nil {
		return err
	}

	block, err := h.keeper.GetBlockByHash(context, blockHash)
	if err != nil {
		return err
	}

	if block.Time() < h.prevBlockTime || int64(block.Time()) >= context.HeaderInfo().Time.Unix() || block.Time() < h.keeper.GetMinBlockTimestamp(context) {
		return fmt.Errorf("symbiotic invalid proposed block")
	}

	if err := h.keeper.SymbioticUpdateValidatorsPower(context, blockHash); err != nil {
		return err
	}

	h.prevBlockTime = block.Time()

	return nil
}
