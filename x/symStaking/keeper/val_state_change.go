package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/math"
	types "cosmossdk.io/x/symStaking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockValidatorUpdates calculates the ValidatorUpdates for the current block
// Called in each EndBlock
func (k Keeper) BlockValidatorUpdates(ctx context.Context) ([]appmodule.ValidatorUpdate, error) {
	// Calculate validator set changes.

	var err error
	for i := 0; i < 3; i++ { // retry 3 times with different providers
		_, err = k.SymbioticUpdateValidatorsPower(ctx)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		return nil, fmt.Errorf("symbiotic error %w", err)
	}

	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	validatorUpdates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	// unbond all mature validators from the unbonding queue
	err = k.UnbondAllMatureValidators(ctx)
	if err != nil {
		return nil, err
	}

	return validatorUpdates, nil
}

// ApplyAndReturnValidatorSetUpdates applies and return accumulated updates to the bonded validator set. Also,
// * Updates the active valset as keyed by LastValidatorPowerKey.
// * Updates the total power as keyed by LastTotalPowerKey.
// * Updates validator status' according to updated powers.
// * Updates the fee pool bonded vs not-bonded tokens.
// * Updates relevant indices.
// It gets called once after genesis, another time maybe after genesis transactions,
// then once at every EndBlock.
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to CometBFT.
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx context.Context) ([]appmodule.ValidatorUpdate, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	maxValidators := params.MaxValidators
	powerReduction := k.PowerReduction(ctx)
	totalPower := math.ZeroInt()

	// Retrieve the last validator set.
	// The persistent set is updated later in this function.
	// (see LastValidatorPowerKey).
	last, err := k.getLastValidatorsByAddr(ctx)
	if err != nil {
		return nil, err
	}

	// Iterate over validators, highest power to lowest.
	iterator, err := k.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var updates []appmodule.ValidatorUpdate
	for count := 0; iterator.Valid() && count < int(maxValidators); iterator.Next() {
		// everything that is iterated in this loop is becoming or already a
		// part of the bonded validator set
		valAddr := sdk.ValAddress(iterator.Value())
		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			return nil, fmt.Errorf("validator record not found for address: %X", valAddr)
		}

		if validator.Jailed {
			return nil, errors.New("should never retrieve a jailed validator from the power store")
		}

		// if we get to a zero-power validator (which we don't bond),
		// there are no more possible bonded validators
		if validator.PotentialConsensusPower(k.PowerReduction(ctx)) == 0 {
			break
		}

		// apply the appropriate state change if necessary
		switch {
		case validator.IsUnbonded():
			validator, err = k.unbondedToBonded(ctx, validator)
			if err != nil {
				return nil, err
			}
		case validator.IsUnbonding():
			validator, err = k.unbondingToBonded(ctx, validator)
			if err != nil {
				return nil, err
			}
		case validator.IsBonded():
			// no state change
		default:
			return nil, errors.New("unexpected validator status")
		}

		// fetch the old power bytes
		valAddrStr, err := k.validatorAddressCodec.BytesToString(valAddr)
		if err != nil {
			return nil, err
		}
		oldPowerBytes, found := last[valAddrStr]
		newPower := validator.ConsensusPower(powerReduction)
		newPowerBytes := k.cdc.MustMarshal(&gogotypes.Int64Value{Value: newPower})

		// update the validator set if power has changed
		if !found || !bytes.Equal(oldPowerBytes, newPowerBytes) {
			updates = append(updates, validator.ModuleValidatorUpdate(powerReduction))
			if err = k.SetLastValidatorPower(ctx, valAddr, newPower); err != nil {
				return nil, err
			}
		}

		delete(last, valAddrStr)
		count++

		totalPower = totalPower.Add(math.NewInt(newPower))
	}

	noLongerBonded, err := sortNoLongerBonded(last, k.validatorAddressCodec)
	if err != nil {
		return nil, err
	}

	for _, valAddrBytes := range noLongerBonded {
		validator, err := k.GetValidator(ctx, sdk.ValAddress(valAddrBytes))
		if err != nil {
			return nil, fmt.Errorf("validator record not found for address: %X", sdk.ValAddress(valAddrBytes))
		}
		validator, err = k.bondedToUnbonding(ctx, validator)
		if err != nil {
			return nil, err
		}
		str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
		if err != nil {
			return nil, err
		}
		if err = k.DeleteLastValidatorPower(ctx, str); err != nil {
			return nil, err
		}

		updates = append(updates, validator.ModuleValidatorUpdateZero())
	}

	// set total power on lookup index if there are any updates
	if len(updates) > 0 {
		if err = k.LastTotalPower.Set(ctx, totalPower); err != nil {
			return nil, err
		}
	}

	return updates, err
}

// Validator state transitions

func (k Keeper) bondedToUnbonding(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsBonded() {
		return types.Validator{}, fmt.Errorf("bad state transition bondedToUnbonding, validator: %v", validator)
	}

	return k.BeginUnbondingValidator(ctx, validator)
}

func (k Keeper) unbondingToBonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonding() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondingToBonded, validator: %v", validator)
	}

	return k.bondValidator(ctx, validator)
}

func (k Keeper) unbondedToBonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonded() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondedToBonded, validator: %v", validator)
	}

	return k.bondValidator(ctx, validator)
}

// UnbondingToUnbonded switches a validator from unbonding state to unbonded state
func (k Keeper) UnbondingToUnbonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonding() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondingToUnbonded, validator: %v", validator)
	}

	return k.completeUnbondingValidator(ctx, validator)
}

// send a validator to jail
func (k Keeper) jailValidator(ctx context.Context, validator types.Validator) error {
	if validator.Jailed {
		return types.ErrValidatorJailed.Wrapf("cannot jail already jailed validator, validator: %v", validator)
	}

	validator.Jailed = true
	if err := k.SetValidator(ctx, validator); err != nil {
		return err
	}

	return k.DeleteValidatorByPowerIndex(ctx, validator)
}

// remove a validator from jail
func (k Keeper) unjailValidator(ctx context.Context, validator types.Validator) error {
	if !validator.Jailed {
		return fmt.Errorf("cannot unjail already unjailed validator, validator: %v", validator)
	}

	validator.Jailed = false
	if err := k.SetValidator(ctx, validator); err != nil {
		return err
	}

	return k.SetValidatorByPowerIndex(ctx, validator)
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	// delete the validator by power index, as the key will change
	if err := k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	validator = validator.UpdateStatus(types.Bonded)

	// save the now bonded validator record to the two referenced stores
	if err := k.SetValidator(ctx, validator); err != nil {
		return validator, err
	}

	if err := k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	// delete from queue if present
	if err := k.DeleteValidatorQueue(ctx, validator); err != nil {
		return validator, err
	}

	// trigger hook
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return validator, err
	}

	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return validator, err
	}

	if err := k.Hooks().AfterValidatorBonded(ctx, consAddr, str); err != nil {
		return validator, err
	}

	return validator, err
}

// BeginUnbondingValidator performs all the store operations for when a validator begins unbonding
func (k Keeper) BeginUnbondingValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return validator, err
	}

	// delete the validator by power index, as the key will change
	if err = k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	// sanity check
	if validator.Status != types.Bonded {
		return validator, fmt.Errorf("should not already be unbonded or unbonding, validator: %v", validator)
	}

	id, err := k.IncrementUnbondingID(ctx)
	if err != nil {
		return validator, err
	}

	validator = validator.UpdateStatus(types.Unbonding)

	headerInfo := k.HeaderService.HeaderInfo(ctx)
	// set the unbonding completion time and completion height appropriately
	validator.UnbondingTime = headerInfo.Time.Add(params.UnbondingTime)
	validator.UnbondingHeight = headerInfo.Height

	validator.UnbondingIds = append(validator.UnbondingIds, id)

	// save the now unbonded validator record and power index
	if err = k.SetValidator(ctx, validator); err != nil {
		return validator, err
	}

	if err = k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	// Adds to unbonding validator queue
	if err = k.InsertUnbondingValidatorQueue(ctx, validator); err != nil {
		return validator, err
	}

	// trigger hook
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return validator, err
	}

	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return validator, err
	}

	if err := k.Hooks().AfterValidatorBeginUnbonding(ctx, consAddr, str); err != nil {
		return validator, err
	}

	if err := k.SetValidatorByUnbondingID(ctx, validator, id); err != nil {
		return validator, err
	}

	if err := k.Hooks().AfterUnbondingInitiated(ctx, id); err != nil {
		return validator, err
	}

	return validator, nil
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) completeUnbondingValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	validator = validator.UpdateStatus(types.Unbonded)
	if err := k.SetValidator(ctx, validator); err != nil {
		return validator, err
	}

	return validator, nil
}

// map of operator bech32-addresses to serialized power
// We use bech32 strings here, because we can't have slices as keys: map[[]byte][]byte
type validatorsByAddr map[string][]byte

// get the last validator set
func (k Keeper) getLastValidatorsByAddr(ctx context.Context) (validatorsByAddr, error) {
	last := make(validatorsByAddr)

	err := k.LastValidatorPower.Walk(ctx, nil, func(key []byte, value gogotypes.Int64Value) (bool, error) {
		valAddrStr, err := k.validatorAddressCodec.BytesToString(key)
		if err != nil {
			return true, err
		}

		intV := value.GetValue()
		bz := k.cdc.MustMarshal(&gogotypes.Int64Value{Value: intV})
		last[valAddrStr] = bz
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return last, nil
}

// given a map of remaining validators to previous bonded power
// returns the list of validators to be unbonded, sorted by operator address
func sortNoLongerBonded(last validatorsByAddr, ac address.Codec) ([][]byte, error) {
	// sort the map keys for determinism
	noLongerBonded := make([][]byte, len(last))
	index := 0

	for valAddrStr := range last {
		valAddrBytes, err := ac.StringToBytes(valAddrStr)
		if err != nil {
			return nil, err
		}
		noLongerBonded[index] = valAddrBytes
		index++
	}
	// sorted by address - order doesn't matter
	sort.SliceStable(noLongerBonded, func(i, j int) bool {
		// -1 means strictly less than
		return bytes.Compare(noLongerBonded[i], noLongerBonded[j]) == -1
	})

	return noLongerBonded, nil
}
