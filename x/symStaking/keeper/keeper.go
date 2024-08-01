package keeper

import (
	"math/rand"
	"os"
	"strings"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"
	"cosmossdk.io/x/symStaking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements ValidatorSet interface
var _ types.ValidatorSet = Keeper{}

func HistoricalInfoCodec(cdc codec.BinaryCodec) collcodec.ValueCodec[types.HistoricalRecord] {
	return collcodec.NewAltValueCodec(codec.CollValue[types.HistoricalRecord](cdc), func(b []byte) (types.HistoricalRecord, error) {
		historicalinfo := types.HistoricalInfo{} //nolint: staticcheck // HistoricalInfo is deprecated
		err := historicalinfo.Unmarshal(b)
		if err != nil {
			return types.HistoricalRecord{}, err
		}

		return types.HistoricalRecord{
			Apphash:        historicalinfo.Header.AppHash,
			Time:           &historicalinfo.Header.Time,
			ValidatorsHash: historicalinfo.Header.NextValidatorsHash,
		}, nil
	})
}

// Keeper of the x/symStaking store
type Keeper struct {
	appmodule.Environment

	cdc                   codec.BinaryCodec
	authKeeper            types.AccountKeeper
	bankKeeper            types.BankKeeper
	hooks                 types.StakingHooks
	authority             string
	validatorAddressCodec addresscodec.Codec
	consensusAddressCodec addresscodec.Codec
	cometInfoService      comet.Service

	beaconAPIURL             string
	ETH_API_URLS             []string
	networkMiddlewareAddress string
	debug                    bool

	Schema collections.Schema

	// HistoricalInfo key: Height | value: HistoricalInfo
	HistoricalInfo collections.Map[uint64, types.HistoricalRecord]
	// LastTotalPower value: LastTotalPower
	LastTotalPower collections.Item[math.Int]
	UnbondingID    collections.Sequence
	// ValidatorByConsensusAddress key: consAddr | value: valAddr
	ValidatorByConsensusAddress collections.Map[sdk.ConsAddress, sdk.ValAddress]
	// UnbondingType key: unbondingID | value: index of UnbondingType
	UnbondingType collections.Map[uint64, uint64]
	// UnbondingIndex key:UnbondingID | value: ubdKey (ubdKey = [UnbondingDelegationKey(Prefix)+len(delAddr)+delAddr+len(valAddr)+valAddr])
	UnbondingIndex collections.Map[uint64, []byte]
	// Validators key: valAddr | value: Validator
	Validators collections.Map[[]byte, types.Validator]
	// ValidatorQueue key: len(timestamp bytes)+timestamp+height | value: ValAddresses
	ValidatorQueue collections.Map[collections.Triple[uint64, time.Time, uint64], types.ValAddresses]
	// LastValidatorPower key: valAddr | value: power(gogotypes.Int64Value())
	LastValidatorPower collections.Map[[]byte, gogotypes.Int64Value]
	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	authority string,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
	cometInfoService comet.Service,
) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)

	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	if validatorAddressCodec == nil || consensusAddressCodec == nil {
		panic("validator and/or consensus address codec are nil")
	}

	beaconAPIURL := os.Getenv("BEACON_API_URL")
	if beaconAPIURL == "" {
		beaconAPIURL = "https://eth-holesky-beacon.public.blastapi.io"
	}

	ethApiUrls := strings.Split(os.Getenv("ETH_API_URLS"), ",")

	if len(ethApiUrls) == 0 {
		ethApiUrls = append(ethApiUrls, "https://endpoints.omniatech.io/v1/eth/holesky/public")
		ethApiUrls = append(ethApiUrls, "https://holesky.drpc.org")
		ethApiUrls = append(ethApiUrls, "https://ethereum-holesky.blockpi.network/v1/rpc/public")
	}

	networkMiddlewareAddress := os.Getenv("MIDDLEWARE_ADDRESS")

	debug := os.Getenv("DEBUG") != ""

	k := &Keeper{
		Environment:              env,
		cdc:                      cdc,
		authKeeper:               ak,
		bankKeeper:               bk,
		hooks:                    nil,
		authority:                authority,
		validatorAddressCodec:    validatorAddressCodec,
		consensusAddressCodec:    consensusAddressCodec,
		cometInfoService:         cometInfoService,
		beaconAPIURL:             beaconAPIURL,
		ETH_API_URLS:             ethApiUrls,
		networkMiddlewareAddress: networkMiddlewareAddress,
		debug:                    debug,
		LastTotalPower:           collections.NewItem(sb, types.LastTotalPowerKey, "last_total_power", sdk.IntValue),
		HistoricalInfo:           collections.NewMap(sb, types.HistoricalInfoKey, "historical_info", collections.Uint64Key, HistoricalInfoCodec(cdc)),
		UnbondingID:              collections.NewSequence(sb, types.UnbondingIDKey, "unbonding_id"),
		ValidatorByConsensusAddress: collections.NewMap(
			sb, types.ValidatorsByConsAddrKey,
			"validator_by_cons_addr",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collcodec.KeyToValueCodec(sdk.ValAddressKey),
		),
		UnbondingType:  collections.NewMap(sb, types.UnbondingTypeKey, "unbonding_type", collections.Uint64Key, collections.Uint64Value),
		UnbondingIndex: collections.NewMap(sb, types.UnbondingIndexKey, "unbonding_index", collections.Uint64Key, collections.BytesValue),
		Validators:     collections.NewMap(sb, types.ValidatorsKey, "validators", sdk.LengthPrefixedBytesKey, codec.CollValue[types.Validator](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		// key format is: 17 | lengthPrefixedBytes(valAddr) | power
		LastValidatorPower: collections.NewMap(sb, types.LastValidatorPowerKey, "last_validator_power", sdk.LengthPrefixedBytesKey, codec.CollValue[gogotypes.Int64Value](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibilitykey format is: 67 | length(timestamp Bytes) | timestamp | height
		// Note: We use 3 keys here because we prefixed time bytes with its length previously and to retain state compatibility we remain to use the same
		ValidatorQueue: collections.NewMap(
			sb, types.ValidatorQueueKey,
			"validator_queue",
			collections.TripleKeyCodec(
				collections.Uint64Key,
				sdk.TimeKey,
				collections.Uint64Key,
			),
			codec.CollValue[types.ValAddresses](cdc),
		),
		// key is: 113 (it's a direct prefix)
		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Hooks gets the hooks for staking *Keeper {
func (k *Keeper) Hooks() types.StakingHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiStakingHooks{}
	}

	return k.hooks
}

// SetHooks sets the validator hooks.  In contrast to other receivers, this method must take a pointer due to nature
// of the hooks interface and SDK start up sequence.
func (k *Keeper) SetHooks(sh types.StakingHooks) {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = sh
}

// GetAuthority returns the x/symStaking module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ValidatorAddressCodec returns the app validator address codec.
func (k Keeper) ValidatorAddressCodec() addresscodec.Codec {
	return k.validatorAddressCodec
}

// ConsensusAddressCodec returns the app consensus address codec.
func (k Keeper) ConsensusAddressCodec() addresscodec.Codec {
	return k.consensusAddressCodec
}

func (k Keeper) GetEthApiUrl() string {
	return k.ETH_API_URLS[rand.Intn(len(k.ETH_API_URLS))]
}
