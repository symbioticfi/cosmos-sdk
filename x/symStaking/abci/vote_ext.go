package abci

import (
	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VoteExtensionHandler struct{}

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(_ sdk.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
		return &abci.ExtendVoteResponse{VoteExtension: nil}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
	}
}
