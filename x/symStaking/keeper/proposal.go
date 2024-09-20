package keeper

import (
	"cosmossdk.io/log"
	"encoding/json"
	"errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProposalHandler struct {
	logger log.Logger
	keeper *Keeper
}

func NewProposalHandler(logger log.Logger, keeper *Keeper) *ProposalHandler {
	return &ProposalHandler{
		logger: logger,
		keeper: keeper,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
		h.logger.Error("PREPARE PROPOSAL")
		proposalTxs := req.Txs

		// get block number
		blockHash, err := h.keeper.GetFinalizedBlockHash(ctx)

		// NOTE: We use stdlib JSON encoding, but an application may choose to use
		// a performant mechanism. This is for demo purposes only.
		bz, err := json.Marshal(blockHash)
		if err != nil {
			h.logger.Error("failed to encode injected vote extension tx", "err", err)
			return nil, errors.New("failed to encode injected vote extension tx")
		}

		// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
		// and store the canonical stake-weighted average prices.
		proposalTxs = append(proposalTxs, bz)

		return &abci.PrepareProposalResponse{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(context sdk.Context, req *abci.FinalizeBlockRequest) error {

		if len(req.Txs) == 0 {
			return nil
		}

		var blockHash string
		if err := json.Unmarshal(req.Txs[0], &blockHash); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return err
		}

		if err := h.keeper.SymbioticUpdateValidatorsPower(context, blockHash); err != nil {
			return err
		}

		return nil
	}
}
