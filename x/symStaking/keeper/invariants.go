package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/x/symStaking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		NonNegativePowerInvariant(k))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return NonNegativePowerInvariant(k)(ctx)
	}
}

// NonNegativePowerInvariant checks that all stored validators have >= 0 power.
func NonNegativePowerInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg    string
			broken bool
		)

		iterator, err := k.ValidatorsPowerStoreIterator(ctx)
		if err != nil {
			panic(err)
		}
		for ; iterator.Valid(); iterator.Next() {
			validator, err := k.GetValidator(ctx, iterator.Value())
			if err != nil {
				panic(fmt.Sprintf("validator record not found for address: %X\n", iterator.Value()))
			}

			powerKey := types.GetValidatorsByPowerIndexKey(validator, k.PowerReduction(ctx), k.ValidatorAddressCodec())

			if !bytes.Equal(iterator.Key(), powerKey) {
				broken = true
				msg += fmt.Sprintf("power store invariance:\n\tvalidator.Power: %v"+
					"\n\tkey should be: %v\n\tkey in store: %v\n",
					validator.GetConsensusPower(k.PowerReduction(ctx)), powerKey, iterator.Key())
			}

			if validator.Tokens.IsNegative() {
				broken = true
				msg += fmt.Sprintf("\tnegative tokens for validator: %v\n", validator)
			}
		}
		iterator.Close()

		return sdk.FormatInvariant(types.ModuleName, "nonnegative power", fmt.Sprintf("found invalid validator powers\n%s", msg)), broken
	}
}
