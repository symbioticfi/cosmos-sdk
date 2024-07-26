package keeper

import (
	"context"
	"fmt"
	"slices"

	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	consensusv1 "cosmossdk.io/x/consensus/types"
	"cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the staking MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if err := msg.Validate(k.validatorAddressCodec); err != nil {
		return nil, err
	}

	minCommRate, err := k.MinCommissionRate(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Commission.Rate.LT(minCommRate) {
		return nil, errorsmod.Wrapf(types.ErrCommissionLTMinRate, "cannot set validator commission to less than minimum rate of %s", minCommRate)
	}

	// check to see if the pubkey or sender has been registered before
	if _, err := k.GetValidator(ctx, valAddr); err == nil {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", msg.Pubkey.GetCachedValue())
	}

	res := consensusv1.QueryParamsResponse{}
	if err := k.QueryRouterService.InvokeTyped(ctx, &consensusv1.QueryParamsRequest{}, &res); err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "failed to query consensus params: %s", err)
	}
	if res.Params.Validator != nil {
		pkType := pk.Type()
		if !slices.Contains(res.Params.Validator.PubKeyTypes, pkType) {
			return nil, errorsmod.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), res.Params.Validator.PubKeyTypes,
			)
		}

		if pkType == sdk.PubKeyEd25519Type && len(pk.Bytes()) != ed25519.PubKeySize {
			return nil, errorsmod.Wrapf(
				types.ErrConsensusPubKeyLenInvalid,
				"got: %d, expected: %d", len(pk.Bytes()), ed25519.PubKeySize,
			)
		}
	}

	err = k.checkConsKeyAlreadyUsed(ctx, pk)
	if err != nil {
		return nil, err
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	validator, err := types.NewValidator(msg.ValidatorAddress, pk, msg.Description)
	if err != nil {
		return nil, err
	}

	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, k.HeaderService.HeaderInfo(ctx).Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		return nil, err
	}

	err = k.SetNewValidatorByPowerIndex(ctx, validator)
	if err != nil {
		return nil, err
	}

	// call the after-creation hook
	if err := k.Hooks().AfterValidatorCreated(ctx, valAddr); err != nil {
		return nil, err
	}

	if err := k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeCreateValidator,
		event.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
	); err != nil {
		return nil, err
	}

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(ctx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	if msg.Description == (types.Description{}) {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(math.LegacyOneDec()) || msg.CommissionRate.IsNegative() {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "commission rate must be between 0 and 1 (inclusive)")
		}

		minCommissionRate, err := k.MinCommissionRate(ctx)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrLogic, err.Error())
		}

		if msg.CommissionRate.LT(minCommissionRate) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commission rate cannot be less than the min commission rate %s", minCommissionRate.String())
		}
	}

	// validator must already be registered
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		if err := k.Hooks().BeforeValidatorModified(ctx, valAddr); err != nil {
			return nil, err
		}

		validator.Commission = commission
	}

	err = k.SetValidator(ctx, validator)
	if err != nil {
		return nil, err
	}

	if err := k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeEditValidator,
		event.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
	); err != nil {
		return nil, err
	}

	return &types.MsgEditValidatorResponse{}, nil
}

// UpdateParams defines a method to perform updation of params exist in x/symStaking module.
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != msg.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	// get previous staking params
	previousParams, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// store params
	if err := k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	// when min commission rate is updated, we need to update the commission rate of all validators
	if !previousParams.MinCommissionRate.Equal(msg.Params.MinCommissionRate) {
		minRate := msg.Params.MinCommissionRate

		vals, err := k.GetAllValidators(ctx)
		if err != nil {
			return nil, err
		}

		for _, val := range vals {
			// set the commission rate to min rate
			if val.Commission.CommissionRates.Rate.LT(minRate) {
				val.Commission.CommissionRates.Rate = minRate
				// set the max rate to minRate if it is less than min rate
				if val.Commission.CommissionRates.MaxRate.LT(minRate) {
					val.Commission.CommissionRates.MaxRate = minRate
				}

				val.Commission.UpdateTime = k.HeaderService.HeaderInfo(ctx).Time
				if err := k.SetValidator(ctx, val); err != nil {
					return nil, fmt.Errorf("failed to set validator after MinCommissionRate param change: %w", err)
				}
			}
		}
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// checkConsKeyAlreadyUsed returns an error if the consensus public key is already used,
// in ConsAddrToValidatorIdentifierMap, OldToNewConsAddrMap, or in the current block (RotationHistory).
func (k msgServer) checkConsKeyAlreadyUsed(ctx context.Context, newConsPubKey cryptotypes.PubKey) error {
	newConsAddr := sdk.ConsAddress(newConsPubKey.Address())

	// checks if NewPubKey is not duplicated on ValidatorsByConsAddr
	_, err := k.Keeper.ValidatorByConsAddr(ctx, newConsAddr)
	if err == nil {
		return types.ErrValidatorPubKeyExists
	}

	return nil
}
