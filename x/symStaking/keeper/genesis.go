package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/symStaking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis sets the parameters for the provided keeper.  For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. In addition, it also sets any delegations found in
// data. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration and delegations
func (k Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) ([]appmodule.ValidatorUpdate, error) {

	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	if sdkCtx, ok := sdk.TryUnwrapSDKContext(ctx); ok {
		// this munging of the context is not necessary for server/v2 code paths, `ok` will be false
		sdkCtx = sdkCtx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay) // TODO: remove this need for WithBlockHeight
		ctx = sdkCtx
	}

	if err := k.Params.Set(ctx, data.Params); err != nil {
		return nil, err
	}

	if err := k.LastTotalPower.Set(ctx, data.LastTotalPower); err != nil {
		return nil, err
	}

	for _, validator := range data.Validators {
		if err := k.SetValidator(ctx, validator); err != nil {
			return nil, err
		}

		// Manually set indices for the first time
		if err := k.SetValidatorByConsAddr(ctx, validator); err != nil {
			return nil, err
		}

		if err := k.SetValidatorByPowerIndex(ctx, validator); err != nil {
			return nil, err
		}

		// Call the creation hook if not exported
		if !data.Exported {
			valbz, err := k.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
			if err != nil {
				return nil, err
			}
			if err := k.Hooks().AfterValidatorCreated(ctx, valbz); err != nil {
				return nil, err
			}
		}

		// update timeslice if necessary
		if validator.IsUnbonding() {
			if err := k.InsertUnbondingValidatorQueue(ctx, validator); err != nil {
				return nil, err
			}
		}

		switch validator.GetStatus() {
		case sdk.Bonded:
		case sdk.Unbonding, sdk.Unbonded:
			continue

		default:
			return nil, fmt.Errorf("invalid validator status: %v", validator.GetStatus())
		}
	}

	// TODO: remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862

	// don't need to run CometBFT updates if we exported
	var moduleValidatorUpdates []appmodule.ValidatorUpdate
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			valAddr, err := k.validatorAddressCodec.StringToBytes(lv.Address)
			if err != nil {
				return nil, err
			}

			err = k.SetLastValidatorPower(ctx, valAddr, lv.Power)
			if err != nil {
				return nil, err
			}

			validator, err := k.GetValidator(ctx, valAddr)
			if err != nil {
				return nil, fmt.Errorf("validator %s not found", lv.Address)
			}

			update := validator.ModuleValidatorUpdate(k.PowerReduction(ctx))
			update.Power = lv.Power // keep the next-val-set offset, use the last power for the first block
			moduleValidatorUpdates = append(moduleValidatorUpdates, update)
		}
	} else {
		var err error

		moduleValidatorUpdates, err = k.BlockValidatorUpdates(ctx)
		if err != nil {
			return nil, err
		}
	}

	return moduleValidatorUpdates, nil
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the params, validators, and bonds found in
// the keeper.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var fnErr error

	var lastValidatorPowers []types.LastValidatorPower

	err := k.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		addrStr, err := k.validatorAddressCodec.BytesToString(addr)
		if err != nil {
			fnErr = err
			return true
		}
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: addrStr, Power: power})
		return false
	})
	if err != nil {
		return nil, err
	}
	if fnErr != nil {
		return nil, fnErr
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	totalPower, err := k.LastTotalPower.Get(ctx)
	if err != nil {
		return nil, err
	}

	allValidators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:              params,
		LastTotalPower:      totalPower,
		LastValidatorPowers: lastValidatorPowers,
		Validators:          allValidators,
		Exported:            true,
	}, nil
}
