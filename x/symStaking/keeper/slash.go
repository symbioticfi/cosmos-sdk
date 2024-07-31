package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// jail a validator
func (k Keeper) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	validator, err := k.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return fmt.Errorf("validator with consensus-Address %s not found", consAddr)
	}
	if err := k.jailValidator(ctx, validator); err != nil {
		return err
	}

	k.Logger.Info("validator jailed", "validator", consAddr)
	return nil
}

// unjail a validator
func (k Keeper) Unjail(ctx context.Context, consAddr sdk.ConsAddress) error {
	validator, err := k.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return fmt.Errorf("validator with consensus-Address %s not found", consAddr)
	}
	if err := k.unjailValidator(ctx, validator); err != nil {
		return err
	}

	k.Logger.Info("validator un-jailed", "validator", consAddr)
	return nil
}
