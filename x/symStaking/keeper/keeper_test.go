package keeper_test

import (
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	consensustypes "cosmossdk.io/x/consensus/types"
	stakingkeeper "cosmossdk.io/x/symStaking/keeper"
	stakingtestutil "cosmossdk.io/x/symStaking/testutil"

	stakingtypes "cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	addresstypes "github.com/cosmos/cosmos-sdk/types/address"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var (
	PKs = simtestutil.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	baseApp       *baseapp.BaseApp
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   stakingtypes.QueryClient
	msgServer     stakingtypes.MsgServer
	key           *storetypes.KVStoreKey
	cdc           codec.Codec
}

func (s *KeeperTestSuite) SetupTest() {
	require := s.Require()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	s.key = key
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.key = key
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})
	s.cdc = encCfg.Codec

	s.baseApp = baseapp.NewBaseApp(
		"staking",
		coretesting.NewNopLogger(),
		testCtx.DB,
		encCfg.TxConfig.TxDecoder(),
	)
	s.baseApp.SetCMS(testCtx.CMS)
	s.baseApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	ctrl := gomock.NewController(s.T())
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	// create consensus keeper
	ck := stakingtestutil.NewMockConsensusKeeper(ctrl)
	ck.EXPECT().Params(gomock.Any(), gomock.Any()).Return(&consensustypes.QueryParamsResponse{
		Params: simtestutil.DefaultConsensusParams,
	}, nil).AnyTimes()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	consensustypes.RegisterQueryServer(queryHelper, ck)

	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)
	env := runtime.NewEnvironment(storeService, coretesting.NewNopLogger(), runtime.EnvWithQueryRouterService(queryHelper.GRPCQueryRouter), runtime.EnvWithMsgRouterService(s.baseApp.MsgServiceRouter()))
	authority, err := accountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(stakingtypes.GovModuleName))
	s.Require().NoError(err)
	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		env,
		accountKeeper,
		bankKeeper,
		authority,
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
		runtime.NewContextAwareCometInfoService(),
	)
	require.NoError(keeper.Params.Set(ctx, stakingtypes.DefaultParams()))

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper

	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)
	s.msgServer = stakingkeeper.NewMsgServerImpl(keeper)
}

func (s *KeeperTestSuite) addressToString(addr []byte) string {
	r, err := s.accountKeeper.AddressCodec().BytesToString(addr)
	s.Require().NoError(err)
	return r
}

func (s *KeeperTestSuite) valAddressToString(addr []byte) string {
	r, err := s.stakingKeeper.ValidatorAddressCodec().BytesToString(addr)
	s.Require().NoError(err)
	return r
}

func (s *KeeperTestSuite) TestParams() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expParams := stakingtypes.DefaultParams()
	// check that the empty keeper loads the default
	resParams, err := keeper.Params.Get(ctx)
	require.NoError(err)
	require.Equal(expParams, resParams)

	expParams.MaxValidators = 555
	expParams.MaxEntries = 111
	require.NoError(keeper.Params.Set(ctx, expParams))
	resParams, err = keeper.Params.Get(ctx)
	require.NoError(err)
	require.True(expParams.Equal(resParams))
}

func (s *KeeperTestSuite) TestLastTotalPower() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expTotalPower := math.NewInt(10 ^ 9)
	require.NoError(keeper.LastTotalPower.Set(ctx, expTotalPower))
	resTotalPower, err := keeper.LastTotalPower.Get(ctx)
	require.NoError(err)
	require.True(expTotalPower.Equal(resTotalPower))
}

// getValidatorKey creates the key for the validator with address
// VALUE: staking/Validator
func getValidatorKey(operatorAddr sdk.ValAddress) []byte {
	validatorsKey := []byte{0x21}
	return append(validatorsKey, addresstypes.MustLengthPrefix(operatorAddr)...)
}

// getLastValidatorPowerKey creates the bonded validator index key for an operator address
func getLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	lastValidatorPowerKey := []byte{0x11}
	return append(lastValidatorPowerKey, addresstypes.MustLengthPrefix(operator)...)
}

// getValidatorQueueKey returns the prefix key used for getting a set of unbonding
// validators whose unbonding completion occurs at the given time and height.
func getValidatorQueueKey(timestamp time.Time, height int64) []byte {
	validatorQueueKey := []byte{0x43}

	heightBz := sdk.Uint64ToBigEndian(uint64(height))
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(validatorQueueKey)

	bz := make([]byte, prefixL+8+timeBzL+8)

	// copy the prefix
	copy(bz[:prefixL], validatorQueueKey)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)

	// copy the encoded height
	copy(bz[prefixL+8+timeBzL:], heightBz)

	return bz
}

func (s *KeeperTestSuite) TestLastTotalPowerMigrationToColls() {
	s.SetupTest()

	_, valAddrs := createValAddrs(100)

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			bz, err := s.cdc.Marshal(&gogotypes.Int64Value{Value: i})
			s.Require().NoError(err)

			s.ctx.KVStore(s.key).Set(getLastValidatorPowerKey(valAddrs[i]), bz)
		},
		"75201270cc94004b2597aed72ac90989de76a3ff3b0081e545f0650d9d4af522",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var intV gogotypes.Int64Value
			intV.Value = i

			err = s.stakingKeeper.LastValidatorPower.Set(s.ctx, valAddrs[i], intV)
			s.Require().NoError(err)
		},
		"75201270cc94004b2597aed72ac90989de76a3ff3b0081e545f0650d9d4af522",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestValidatorsMigrationToColls() {
	s.SetupTest()
	pkAny, err := codectypes.NewAnyWithValue(PKs[0])
	s.Require().NoError(err)

	_, valAddrs := createValAddrs(100)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			val := stakingtypes.Validator{
				OperatorAddress: s.valAddressToString(valAddrs[i]),
				ConsensusPubkey: pkAny,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.DefaultPowerReduction,
				Description:     stakingtypes.Description{},
				UnbondingHeight: int64(0),
				UnbondingTime:   time.Unix(0, 0).UTC(),
			}
			valBz := s.cdc.MustMarshal(&val)
			// legacy Set method
			s.ctx.KVStore(s.key).Set(getValidatorKey(valAddrs[i]), valBz)
		},
		"5d54b1aaba92af47582e57fb61b695564c42a68686fc695c10d690981681248f",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			val := stakingtypes.Validator{
				OperatorAddress: s.valAddressToString(valAddrs[i]),
				ConsensusPubkey: pkAny,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.DefaultPowerReduction,
				Description:     stakingtypes.Description{},
				UnbondingHeight: int64(0),
				UnbondingTime:   time.Unix(0, 0).UTC(),
			}

			err := s.stakingKeeper.SetValidator(s.ctx, val)
			s.Require().NoError(err)
		},
		"5d54b1aaba92af47582e57fb61b695564c42a68686fc695c10d690981681248f",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestValidatorQueueMigrationToColls() {
	s.SetupTest()
	_, valAddrs := createValAddrs(100)
	endTime := time.Unix(0, 0).UTC()
	endHeight := int64(10)
	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var addrs []string
			addrs = append(addrs, s.valAddressToString(valAddrs[i]))
			bz, err := s.cdc.Marshal(&stakingtypes.ValAddresses{Addresses: addrs})
			s.Require().NoError(err)

			// legacy Set method
			s.ctx.KVStore(s.key).Set(getValidatorQueueKey(endTime, endHeight), bz)
		},
		"bdd568b910d8ab2e74511844f223edf21be6dacea7e2535cebf4683c81a3b591",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var addrs []string
			addrs = append(addrs, s.valAddressToString(valAddrs[i]))

			err := s.stakingKeeper.SetUnbondingValidatorsQueue(s.ctx, endTime, endHeight, addrs)
			s.Require().NoError(err)
		},
		"bdd568b910d8ab2e74511844f223edf21be6dacea7e2535cebf4683c81a3b591",
	)
	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func createValAddrs(count int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := simtestutil.CreateIncrementalAccounts(count)
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	return addrs, valAddrs
}
