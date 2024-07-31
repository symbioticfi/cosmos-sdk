package keeper_test

import (
	"cosmossdk.io/x/symStaking/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// tests Jail, Unjail
func (s *KeeperTestSuite) TestRevocation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	consAddr := sdk.ConsAddress(PKs[0].Address())
	validator := testutil.NewValidator(s.T(), valAddr, PKs[0])

	// initial state
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))
	val, err := keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.False(val.IsJailed())

	// test jail
	require.NoError(keeper.Jail(ctx, consAddr))
	val, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.True(val.IsJailed())

	// test unjail
	require.NoError(keeper.Unjail(ctx, consAddr))
	val, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.False(val.IsJailed())
}
