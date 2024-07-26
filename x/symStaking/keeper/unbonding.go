package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/symStaking/types"
)

// IncrementUnbondingID increments and returns a unique ID for an unbonding operation
func (k Keeper) IncrementUnbondingID(ctx context.Context) (unbondingID uint64, err error) {
	unbondingID, err = k.UnbondingID.Next(ctx)
	if err != nil {
		return 0, err
	}
	unbondingID++

	return unbondingID, err
}

// DeleteUnbondingIndex removes a mapping from UnbondingId to unbonding operation
func (k Keeper) DeleteUnbondingIndex(ctx context.Context, id uint64) error {
	return k.UnbondingIndex.Remove(ctx, id)
}

// GetUnbondingType returns the enum type of unbonding which is any of
// {UnbondingDelegation | Redelegation | ValidatorUnbonding}
func (k Keeper) GetUnbondingType(ctx context.Context, id uint64) (unbondingType types.UnbondingType, err error) {
	ubdType, err := k.UnbondingType.Get(ctx, id)
	if errors.Is(err, collections.ErrNotFound) {
		return unbondingType, types.ErrNoUnbondingType
	}
	return types.UnbondingType(ubdType), err
}

// SetUnbondingType sets the enum type of unbonding which is any of
// {UnbondingDelegation | Redelegation | ValidatorUnbonding}
func (k Keeper) SetUnbondingType(ctx context.Context, id uint64, unbondingType types.UnbondingType) error {
	return k.UnbondingType.Set(ctx, id, uint64(unbondingType))
}

// GetValidatorByUnbondingID returns the validator that is unbonding with a certain unbonding op ID
func (k Keeper) GetValidatorByUnbondingID(ctx context.Context, id uint64) (val types.Validator, err error) {
	store := k.KVStoreService.OpenKVStore(ctx)

	valKey, err := k.UnbondingIndex.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Validator{}, types.ErrNoValidatorFound
		}
		return types.Validator{}, err
	}

	if valKey == nil {
		return types.Validator{}, types.ErrNoValidatorFound
	}

	value, err := store.Get(valKey)
	if err != nil {
		return types.Validator{}, err
	}

	if value == nil {
		return types.Validator{}, types.ErrNoValidatorFound
	}

	val, err = types.UnmarshalValidator(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.Validator{}, err
	}

	return val, nil
}

// SetValidatorByUnbondingID sets an index to look up a Validator by the unbondingID corresponding to its current unbonding
// Note, it does not set the validator itself, use SetValidator(ctx, val) for that
func (k Keeper) SetValidatorByUnbondingID(ctx context.Context, val types.Validator, id uint64) error {
	valAddr, err := k.validatorAddressCodec.StringToBytes(val.OperatorAddress)
	if err != nil {
		return err
	}

	valKey := types.GetValidatorKey(valAddr)
	if err = k.UnbondingIndex.Set(ctx, id, valKey); err != nil {
		return err
	}

	// Set unbonding type so that we know how to deserialize it later
	return k.SetUnbondingType(ctx, id, types.UnbondingType_ValidatorUnbonding)
}

// UnbondingCanComplete allows a stopped unbonding operation, such as an
// unbonding delegation, a redelegation, or a validator unbonding to complete.
// In order for the unbonding operation with `id` to eventually complete, every call
// to PutUnbondingOnHold(id) must be matched by a call to UnbondingCanComplete(id).
func (k Keeper) UnbondingCanComplete(ctx context.Context, id uint64) error {
	unbondingType, err := k.GetUnbondingType(ctx, id)
	if err != nil {
		return err
	}

	switch unbondingType {
	case types.UnbondingType_ValidatorUnbonding:
		if err := k.validatorUnbondingCanComplete(ctx, id); err != nil {
			return err
		}
	default:
		return types.ErrUnbondingNotFound
	}

	return nil
}

func (k Keeper) validatorUnbondingCanComplete(ctx context.Context, id uint64) error {
	val, err := k.GetValidatorByUnbondingID(ctx, id)
	if err != nil {
		return err
	}

	if val.UnbondingOnHoldRefCount <= 0 {
		return errorsmod.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"val(%s), expecting UnbondingOnHoldRefCount > 0, got %T",
			val.OperatorAddress, val.UnbondingOnHoldRefCount,
		)
	}
	val.UnbondingOnHoldRefCount--
	return k.SetValidator(ctx, val)
}

// PutUnbondingOnHold allows an external module to stop an unbonding operation,
// such as an unbonding delegation, a redelegation, or a validator unbonding.
// In order for the unbonding operation with `id` to eventually complete, every call
// to PutUnbondingOnHold(id) must be matched by a call to UnbondingCanComplete(id).
func (k Keeper) PutUnbondingOnHold(ctx context.Context, id uint64) error {
	unbondingType, err := k.GetUnbondingType(ctx, id)
	if err != nil {
		return err
	}
	switch unbondingType {
	case types.UnbondingType_ValidatorUnbonding:
		if err := k.putValidatorOnHold(ctx, id); err != nil {
			return err
		}
	default:
		return types.ErrUnbondingNotFound
	}

	return nil
}

func (k Keeper) putValidatorOnHold(ctx context.Context, id uint64) error {
	val, err := k.GetValidatorByUnbondingID(ctx, id)
	if err != nil {
		return err
	}

	val.UnbondingOnHoldRefCount++
	return k.SetValidator(ctx, val)
}
