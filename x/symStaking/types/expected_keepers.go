package types

import (
	"context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	consensustypes "cosmossdk.io/x/consensus/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	AddressCodec() address.Codec

	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI // only used for simulation

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	GetSupply(ctx context.Context, denom string) sdk.Coin

	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderPool, recipientPool string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	BurnCoins(context.Context, []byte, sdk.Coins) error
	IsSendEnabledDenom(ctx context.Context, denom string) bool
}

// ValidatorSet expected properties for the set of all validators (noalias)
type ValidatorSet interface {
	// iterate through validators by operator address, execute func for each validator
	IterateValidators(context.Context,
		func(index int64, validator Validator) (stop bool)) error

	// iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(context.Context,
		func(index int64, validator Validator) (stop bool)) error

	Validator(context.Context, sdk.ValAddress) (Validator, error)            // get a particular validator by operator address
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (Validator, error) // get a particular validator by consensus address

	Jail(context.Context, sdk.ConsAddress) error   // jail a validator
	Unjail(context.Context, sdk.ConsAddress) error // unjail a validator

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(context.Context) (uint32, error)

	// GetPubKeyByConsAddr returns the consensus public key for a validator. Used in vote
	// extension validation.
	GetPubKeyByConsAddr(context.Context, sdk.ConsAddress) (cryptotypes.PubKey, error)

	TotalBondedTokens(context.Context) (math.Int, error) // total bonded tokens within the validator set
}

// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error                           // Must be called when a validator is created
	BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error                         // Must be called when a validator's state changes
	AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator is deleted

	AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator begins unbonding

	BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error
	AfterUnbondingInitiated(ctx context.Context, id uint64) error
}

// StakingHooksWrapper is a wrapper for modules to inject StakingHooks using depinject.
type StakingHooksWrapper struct{ StakingHooks }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (StakingHooksWrapper) IsOnePerModuleType() {}

type ConsensusKeeper interface {
	Params(context.Context, *consensustypes.QueryParamsRequest) (*consensustypes.QueryParamsResponse, error)
	GetCometInfo(context.Context, *consensustypes.QueryGetCometInfoRequest) (*consensustypes.QueryGetCometInfoResponse, error)
}
