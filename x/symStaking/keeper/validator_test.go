package keeper_test

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	stakingkeeper "cosmossdk.io/x/symStaking/keeper"
	"cosmossdk.io/x/symStaking/testutil"
	stakingtypes "cosmossdk.io/x/symStaking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) applyValidatorSetUpdates(ctx sdk.Context, keeper *stakingkeeper.Keeper, expectedUpdatesLen int) []appmodule.ValidatorUpdate {
	updates, err := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	s.Require().NoError(err)
	if expectedUpdatesLen >= 0 {
		s.Require().Equal(expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}

func (s *KeeperTestSuite) TestValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := keeper.TokensFromConsensusPower(ctx, 10)

	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator = validator.AddTokens(valTokens)
	require.Equal(stakingtypes.Unbonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.SetValidatorByPowerIndex(ctx, validator))
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	// ensure update
	updates := s.applyValidatorSetUpdates(ctx, keeper, 1)
	validator, err := keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(validator.ModuleValidatorUpdate(keeper.PowerReduction(ctx)), updates[0])

	// after the save the validator should be bonded
	require.Equal(stakingtypes.Bonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)

	// check each store for being saved
	consAddr, err := validator.GetConsAddr()
	require.NoError(err)
	resVal, err := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.NoError(err)
	require.True(validator.MinEqual(&resVal))

	resVals, err := keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	resVals, err = keeper.GetBondedValidatorsByPower(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	allVals, err := keeper.GetAllValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(allVals))

	// check the last validator power
	power := int64(100)
	require.NoError(keeper.SetLastValidatorPower(ctx, valAddr, power))
	resPower, err := keeper.GetLastValidatorPower(ctx, valAddr)
	require.NoError(err)
	require.Equal(power, resPower)
	require.NoError(keeper.DeleteLastValidatorPower(ctx, valAddr))
	resPower, err = keeper.GetLastValidatorPower(ctx, valAddr)
	require.Error(err, collections.ErrNotFound)
	require.Equal(int64(0), resPower)
}

func (s *KeeperTestSuite) TestGetLastValidators() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()
	params, err := keeper.Params.Get(ctx)
	require.NoError(err)
	params.MaxValidators = 50
	require.NoError(keeper.Params.Set(ctx, params))
	// construct 50 validators all with equal power of 100
	var validators [50]stakingtypes.Validator
	for i := 0; i < 50; i++ {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		validators[i].Status = stakingtypes.Unbonded
		validators[i].Tokens = math.ZeroInt()
		tokens := keeper.TokensFromConsensusPower(ctx, 100)
		validators[i] = validators[i].AddTokens(tokens)
		require.Equal(keeper.TokensFromConsensusPower(ctx, 100), validators[i].Tokens)
		validators[i] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[i], true)
		require.NoError(keeper.SetValidatorByConsAddr(ctx, validators[i]))
		resVal, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[i].Address().Bytes()))
		require.NoError(err)
		require.True(validators[i].MinEqual(&resVal))
	}

	res, err := keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Len(res, 50)

	// reduce max validators to 30 and ensure we only get 30 back
	params.MaxValidators = 30
	require.NoError(keeper.Params.Set(ctx, params))

	res, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Len(res, 30)
}

// This function tests UpdateValidator, GetValidator, GetLastValidators, RemoveValidator
func (s *KeeperTestSuite) TestValidatorBasics() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// construct the validators
	var validators [3]stakingtypes.Validator
	powers := []int64{9, 8, 7}
	for i, power := range powers {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		validators[i].Status = stakingtypes.Unbonded
		validators[i].Tokens = math.ZeroInt()
		tokens := keeper.TokensFromConsensusPower(ctx, power)

		validators[i] = validators[i].AddTokens(tokens)
	}

	require.Equal(keeper.TokensFromConsensusPower(ctx, 9), validators[0].Tokens)
	require.Equal(keeper.TokensFromConsensusPower(ctx, 8), validators[1].Tokens)
	require.Equal(keeper.TokensFromConsensusPower(ctx, 7), validators[2].Tokens)

	// check the empty keeper first
	_, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
	resVals, err := keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Zero(len(resVals))

	resVals, err = keeper.GetValidators(ctx, 2)
	require.NoError(err)
	require.Len(resVals, 0)

	// set and retrieve a record
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], true)
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validators[0]))
	resVal, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	// retrieve from consensus
	resVal, err = keeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(PKs[0].Address()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))
	resVal, err = keeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validators[0].MinEqual(&resVals[0]))
	require.Equal(stakingtypes.Bonded, validators[0].Status)
	require.True(keeper.TokensFromConsensusPower(ctx, 9).Equal(validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = stakingtypes.Bonded
	validators[0].Tokens = keeper.TokensFromConsensusPower(ctx, 10)
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], true)
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validators[0].MinEqual(&resVals[0]))

	// add other validators
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], true)
	validators[2] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[2], true)
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[1].Address().Bytes()))
	require.NoError(err)
	require.True(validators[1].MinEqual(&resVal))
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[2].Address().Bytes()))
	require.NoError(err)
	require.True(validators[2].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(3, len(resVals))

	// remove a record

	bz, err := keeper.ValidatorAddressCodec().StringToBytes(validators[1].GetOperator())
	require.NoError(err)

	// shouldn't be able to remove if status is not unbonded
	require.EqualError(keeper.RemoveValidator(ctx, bz), "cannot call RemoveValidator on bonded or unbonding validators: failed to remove validator")

	// shouldn't be able to remove if there are still tokens left
	validators[1].Status = stakingtypes.Unbonded
	require.NoError(keeper.SetValidator(ctx, validators[1]))
	require.EqualError(keeper.RemoveValidator(ctx, bz), "attempting to remove a validator which still contains tokens: failed to remove validator")

	validators[1].Tokens = math.ZeroInt()                    // ...remove all tokens
	require.NoError(keeper.SetValidator(ctx, validators[1])) // ...set the validator
	require.NoError(keeper.RemoveValidator(ctx, bz))         // Now it can be removed.
	_, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[1].Address().Bytes()))
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
}

func (s *KeeperTestSuite) TestUpdateValidatorByPowerIndex() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := keeper.TokensFromConsensusPower(ctx, 100)

	// add a validator
	validator := testutil.NewValidator(s.T(), valAddr, PKs[0])
	validator = validator.AddTokens(valTokens)
	require.Equal(stakingtypes.Unbonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)

	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	validator, err := keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(valTokens, validator.Tokens)

	power := stakingtypes.GetValidatorsByPowerIndexKey(validator, keeper.PowerReduction(ctx), keeper.ValidatorAddressCodec())
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	// burn half the delegator shares
	validator = validator.RemoveTokens(keeper.TokensFromConsensusPower(ctx, 50))
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	validator, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)

	power = stakingtypes.GetValidatorsByPowerIndexKey(validator, keeper.PowerReduction(ctx), keeper.ValidatorAddressCodec())
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	// set new validator by power index
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	require.False(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))
	require.NoError(keeper.SetNewValidatorByPowerIndex(ctx, validator))
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))
}

func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdatesPowerDecrease() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	powers := []int64{100, 100}
	var validators [2]stakingtypes.Validator

	for i, power := range powers {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		tokens := keeper.TokensFromConsensusPower(ctx, power)
		validators[i] = validators[i].AddTokens(tokens)
	}

	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], false)
	s.applyValidatorSetUpdates(ctx, keeper, 2)

	// check initial power
	require.Equal(int64(100), validators[0].GetConsensusPower(keeper.PowerReduction(ctx)))
	require.Equal(int64(100), validators[1].GetConsensusPower(keeper.PowerReduction(ctx)))

	validators[0] = validators[0].RemoveTokens(keeper.TokensFromConsensusPower(ctx, 20))
	validators[1] = validators[1].RemoveTokens(keeper.TokensFromConsensusPower(ctx, 30))
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], false)

	// power has changed
	require.Equal(int64(80), validators[0].GetConsensusPower(keeper.PowerReduction(ctx)))
	require.Equal(int64(70), validators[1].GetConsensusPower(keeper.PowerReduction(ctx)))

	// CometBFT updates should reflect power change
	updates := s.applyValidatorSetUpdates(ctx, keeper, 2)
	require.Equal(validators[0].ModuleValidatorUpdate(keeper.PowerReduction(ctx)), updates[0])
	require.Equal(validators[1].ModuleValidatorUpdate(keeper.PowerReduction(ctx)), updates[1])
}

func (s *KeeperTestSuite) TestValidatorToken() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	addTokens := keeper.TokensFromConsensusPower(ctx, 10)
	delTokens := keeper.TokensFromConsensusPower(ctx, 5)

	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator, err := keeper.AddValidatorTokens(ctx, validator, addTokens)
	require.NoError(err)
	require.Equal(addTokens, validator.Tokens)
	validator, _ = keeper.GetValidator(ctx, valAddr)

	_, err = keeper.RemoveValidatorTokens(ctx, validator, delTokens)
	require.NoError(err)
	validator, _ = keeper.GetValidator(ctx, valAddr)
	require.Equal(delTokens, validator.Tokens)

	_, err = keeper.RemoveValidatorTokens(ctx, validator, delTokens)
	require.NoError(err)
	validator, _ = keeper.GetValidator(ctx, valAddr)
	require.True(validator.Tokens.IsZero())
}

// TestUnbondingValidator tests the functionality of unbonding a validator.
func (s *KeeperTestSuite) TestUnbondingValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	addTokens := keeper.TokensFromConsensusPower(ctx, 10)

	// set unbonding validator
	endTime := time.Now()
	endHeight := ctx.HeaderInfo().Height + 10
	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, endTime, endHeight, []string{s.valAddressToString(valAddr)}))

	resVals, err := keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.Equal(s.valAddressToString(valAddr), resVals[0])

	// add another unbonding validator
	valAddr1 := sdk.ValAddress(PKs[1].Address().Bytes())
	validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[1])
	validator1.UnbondingHeight = endHeight
	validator1.UnbondingTime = endTime
	require.NoError(keeper.InsertUnbondingValidatorQueue(ctx, validator1))

	resVals, err = keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(2, len(resVals))

	// delete unbonding validator from the queue
	require.NoError(keeper.DeleteValidatorQueue(ctx, validator1))
	resVals, err = keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.Equal(s.valAddressToString(valAddr), resVals[0])

	// check unbonding mature validators
	ctx = ctx.WithHeaderInfo(header.Info{Height: endHeight, Time: endTime})
	err = keeper.UnbondAllMatureValidators(ctx)
	require.EqualError(err, "validator in the unbonding queue was not found: validator does not exist")

	require.NoError(keeper.SetValidator(ctx, validator))
	ctx = ctx.WithHeaderInfo(header.Info{Height: endHeight, Time: endTime})

	err = keeper.UnbondAllMatureValidators(ctx)
	require.EqualError(err, "unexpected validator in unbonding queue; status was not unbonding")

	validator.Status = stakingtypes.Unbonding
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.UnbondAllMatureValidators(ctx))
	validator, err = keeper.GetValidator(ctx, valAddr)
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)

	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, endTime, endHeight, []string{s.valAddressToString(valAddr)}))
	validator = testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator = validator.AddTokens(addTokens)
	validator.Status = stakingtypes.Unbonding
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.UnbondAllMatureValidators(ctx))
	validator, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, validator.Status)
}
