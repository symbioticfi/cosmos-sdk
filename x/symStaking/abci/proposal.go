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
)

const SYMBIOTIC_SYNC_PERIOD = 10

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
		proposalTxs := req.Txs

		if req.Height%SYMBIOTIC_SYNC_PERIOD != 0 {
			return &abci.PrepareProposalResponse{
				Txs: proposalTxs,
			}, nil
		}

		blockHash, err := h.keeper.GetFinalizedBlockHash(ctx)
		if err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		// NOTE: We use stdlib JSON encoding, but an application may choose to use
		// a performant mechanism. This is for demo purposes only.
		bz, err := json.Marshal(blockHash)
		if err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, errors.New("failed to encode injected vote extension tx")))
		}

		// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
		// and store the canonical stake-weighted average prices.
		proposalTxs = append([][]byte{bz}, proposalTxs...)

		return &abci.PrepareProposalResponse{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(context sdk.Context, req *abci.FinalizeBlockRequest) error {

		if req.Height%SYMBIOTIC_SYNC_PERIOD != 0 || len(req.Txs) == 0 {
			return nil
		}

		var blockHash string
		if err := json.Unmarshal(req.Txs[0], &blockHash); err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		block, err := h.keeper.GetBlockByHash(context, blockHash)
		if err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		if block.Time() < h.prevBlockTime || int64(block.Time()) >= context.HeaderInfo().Time.Unix() || block.Time() < h.keeper.GetMinBlockTimestamp(context) {
			panic(errors.Join(types.ErrSymbioticValUpdate, fmt.Errorf("symbiotic invalid proposed block")))
		}

		if err := h.keeper.SymbioticUpdateValidatorsPower(context, blockHash); err != nil {
			panic(errors.Join(types.ErrSymbioticValUpdate, err))
		}

		h.prevBlockTime = block.Time()

		return nil
	}
}
