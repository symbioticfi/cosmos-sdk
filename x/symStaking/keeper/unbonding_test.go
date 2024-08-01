package keeper_test

import (
	"cosmossdk.io/x/symStaking/testutil"
	"cosmossdk.io/x/symStaking/types"
)

func (s *KeeperTestSuite) TestIncrementUnbondingID() {
	for i := 1; i < 10; i++ {
		id, err := s.stakingKeeper.IncrementUnbondingID(s.ctx)
		s.Require().NoError(err)
		s.Require().Equal(uint64(i), id)
	}
}

func (s *KeeperTestSuite) TestUnbondingTypeAccessors() {
	require := s.Require()
	cases := []struct {
		exists   bool
		name     string
		expected types.UnbondingType
	}{
		{
			name:   "not existing",
			exists: false,
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists {
				require.NoError(s.stakingKeeper.SetUnbondingType(s.ctx, uint64(i), tc.expected))
			}

			unbondingType, err := s.stakingKeeper.GetUnbondingType(s.ctx, uint64(i))
			if tc.exists {
				require.NoError(err)
				require.Equal(tc.expected, unbondingType)
			} else {
				require.ErrorIs(err, types.ErrNoUnbondingType)
			}
		})
	}
}
func (s *KeeperTestSuite) TestValidatorByUnbondingIDAccessors() {
	_, valAddrs := createValAddrs(3)
	require := s.Require()

	type exists struct {
		setValidator              bool
		setValidatorByUnbondingID bool
	}

	cases := []struct {
		exists    exists
		name      string
		validator types.Validator
	}{
		{
			name:      "existing 1",
			exists:    exists{true, true},
			validator: testutil.NewValidator(s.T(), valAddrs[0], PKs[0]),
		},
		{
			name:      "not existing 1",
			exists:    exists{false, true},
			validator: testutil.NewValidator(s.T(), valAddrs[1], PKs[1]),
		},
		{
			name:      "not existing 2",
			exists:    exists{false, false},
			validator: testutil.NewValidator(s.T(), valAddrs[2], PKs[0]),
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists.setValidator {
				require.NoError(s.stakingKeeper.SetValidator(s.ctx, tc.validator))
			}

			if tc.exists.setValidatorByUnbondingID {
				require.NoError(s.stakingKeeper.SetValidatorByUnbondingID(s.ctx, tc.validator, uint64(i)))
			}

			val, err := s.stakingKeeper.GetValidatorByUnbondingID(s.ctx, uint64(i))
			if tc.exists.setValidator && tc.exists.setValidatorByUnbondingID {
				require.NoError(err)
				require.Equal(tc.validator, val)
			} else {
				require.ErrorIs(err, types.ErrNoValidatorFound)
			}
		})
	}
}
