package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/x/symSlash/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Unjail calls the staking Unjail function to unjail a validator if the
// jailed period has concluded
func (k Keeper) Unjail(ctx context.Context, validatorAddr sdk.ValAddress) error {
	validator, err := k.sk.Validator(ctx, validatorAddr)
	if err != nil {
		return err
	}

	tokens := validator.GetTokens()
	if tokens.LT(sdk.DefaultPowerReduction) {
		return errors.Wrapf(
			types.ErrSelfDelegationTooLowToUnjail, "%s less than %s", tokens, sdk.DefaultPowerReduction,
		)
	}

	// cannot be unjailed if not jailed
	if !validator.IsJailed() {
		return types.ErrValidatorNotJailed
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}
	// If the validator has a ValidatorSigningInfo object that signals that the
	// validator was bonded and so we must check that the validator is not tombstoned
	// and can be unjailed at the current block.
	//
	// A validator that is jailed but has no ValidatorSigningInfo object signals
	// that the validator was never bonded and must've been jailed due to falling
	// below their minimum self-delegation. The validator can unjail at any point
	// assuming they've now bonded above their minimum self-delegation.
	info, err := k.ValidatorSigningInfo.Get(ctx, consAddr)
	if err == nil {
		// cannot be unjailed if tombstoned
		if info.Tombstoned {
			return types.ErrValidatorJailed
		}

		if k.HeaderService.HeaderInfo(ctx).Time.Before(info.JailedUntil) {
			return types.ErrValidatorJailed
		}
	}

	return k.sk.Unjail(ctx, consAddr)
}
