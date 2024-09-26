package types

import (
	"context"
	"cosmossdk.io/core/appmodule"
	bankexported "cosmossdk.io/x/bank/exported"
	stakingtypes "cosmossdk.io/x/symStaking/types"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StakingKeeper defines the expected staking keeper (noalias)
type StakingKeeper interface {
	CacheBlockHash(ctx context.Context, blockHash stakingtypes.CachedBlockHash) error
	BlockValidatorUpdates(ctx context.Context) ([]appmodule.ValidatorUpdate, error)
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
}

// GenesisAccountsIterator defines the expected iterating genesis accounts object (noalias)
type GenesisAccountsIterator interface {
	IterateGenesisAccounts(
		cdc *codec.LegacyAmino,
		appGenesis map[string]json.RawMessage,
		cb func(sdk.AccountI) (stop bool),
	)
}

// GenesisBalancesIterator defines the expected iterating genesis balances object (noalias)
type GenesisBalancesIterator interface {
	IterateGenesisBalances(
		cdc codec.JSONCodec,
		appGenesis map[string]json.RawMessage,
		cb func(bankexported.GenesisBalance) (stop bool),
	)
}
