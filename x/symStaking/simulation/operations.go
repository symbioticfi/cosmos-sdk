package simulation

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/math"
	"cosmossdk.io/x/symStaking/keeper"
	"cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgCreateValidator int = 100
	DefaultWeightMsgEditValidator   int = 5

	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator   = "op_weight_msg_edit_validator"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator int
		weightMsgEditValidator   int
	)

	appParams.GetOrGenerate(OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil, func(_ *rand.Rand) {
		weightMsgCreateValidator = DefaultWeightMsgCreateValidator
	})

	appParams.GetOrGenerate(OpWeightMsgEditValidator, &weightMsgEditValidator, nil, func(_ *rand.Rand) {
		weightMsgEditValidator = DefaultWeightMsgEditValidator
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditValidator,
			SimulateMsgEditValidator(txGen, ak, bk, k),
		)}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgCreateValidator{})

		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, err := k.GetValidator(ctx, address)
		if err == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator already exists"), nil, nil
		}

		consPubKey := sdk.GetConsAddress(simAccount.ConsKey.PubKey())
		_, err = k.GetValidatorByConsAddr(ctx, consPubKey)
		if err == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "cons key already used"), nil, nil
		}

		denom, err := k.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "bond denom not found"), nil, err
		}

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "balance is negative"), nil, nil
		}

		amount, err := simtypes.RandPositiveInt(r, balance)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(selfDelegation)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate fees"), nil, err
			}
		}

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		maxCommission := math.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
		commission := types.NewCommissionRates(
			simtypes.RandomDecAmount(r, maxCommission),
			maxCommission,
			simtypes.RandomDecAmount(r, maxCommission),
		)

		addr, err := k.ValidatorAddressCodec().BytesToString(address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate validator address"), nil, err
		}

		msg, err := types.NewMsgCreateValidator(addr, simAccount.ConsKey.PubKey(), description, commission)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "unable to create CreateValidator message"), nil, err
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgEditValidator generates a MsgEditValidator with random values
func SimulateMsgEditValidator(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgEditValidator{})

		vals, err := k.GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}

		if len(vals) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "number of validators equal zero"), nil, nil
		}

		val, ok := testutil.RandSliceElem(r, vals)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pick a validator"), nil, nil
		}

		address := val.GetOperator()
		newCommissionRate := simtypes.RandomDecAmount(r, val.Commission.MaxRate)

		if err := val.Commission.ValidateNewRate(newCommissionRate, ctx.HeaderInfo().Time); err != nil {
			// skip as the commission is invalid
			return simtypes.NoOpMsg(types.ModuleName, msgType, "invalid commission rate"), nil, nil
		}

		bz, err := k.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting validator address bytes"), nil, err
		}

		simAccount, found := simtypes.FindAccount(accs, sdk.AccAddress(bz))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to find account"), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgEditValidator(address, &newCommissionRate, description)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
