package keeper_test

import "log"

// tests Jail, Unjail
func (s *KeeperTestSuite) TestSymbioticChange() {
	ctx, keeper := s.ctx, s.stakingKeeper
	_, err := keeper.SymbioticUpdateValidatorsPower(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
