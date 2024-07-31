package types

import (
	"context"

	"cosmossdk.io/core/address"
	stakingtypes "cosmossdk.io/x/symStaking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper expected account keeper
type AccountKeeper interface {
	AddressCodec() address.Codec
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(context.Context,
		func(index int64, validator stakingtypes.Validator) (stop bool)) error

	Validator(context.Context, sdk.ValAddress) (stakingtypes.Validator, error)            // get a particular validator by operator address
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.Validator, error) // get a particular validator by consensus address

	Jail(context.Context, sdk.ConsAddress) error   // jail a validator
	Unjail(context.Context, sdk.ConsAddress) error // unjail a validator

	GetAllValidators(ctx context.Context) ([]stakingtypes.Validator, error)

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(context.Context) (uint32, error)

	// IsValidatorJailed returns if the validator is jailed.
	IsValidatorJailed(ctx context.Context, addr sdk.ConsAddress) (bool, error)
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error                           // Must be called when a validator is created
	BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error                         // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator is deleted

	AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator begins unbonding
}
