package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/x/symSlash/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper of the slashing store
type Keeper struct {
	appmodule.Environment

	cdc codec.BinaryCodec
	// deprecated!
	legacyAmino *codec.LegacyAmino
	sk          types.StakingKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
	Schema    collections.Schema
	Params    collections.Item[types.Params]
	// ValidatorSigningInfo key: ConsAddr | value: ValidatorSigningInfo
	ValidatorSigningInfo collections.Map[sdk.ConsAddress, types.ValidatorSigningInfo]
	// AddrPubkeyRelation key: address | value: PubKey
	AddrPubkeyRelation collections.Map[[]byte, cryptotypes.PubKey]
	// ValidatorMissedBlockBitmap key: ConsAddr | value: byte key for a validator's missed block bitmap chunk
	ValidatorMissedBlockBitmap collections.Map[collections.Pair[[]byte, uint64], []byte]
}

// NewKeeper creates a slashing keeper
func NewKeeper(environment appmodule.Environment, cdc codec.BinaryCodec, legacyAmino *codec.LegacyAmino, sk types.StakingKeeper, authority string) Keeper {
	sb := collections.NewSchemaBuilder(environment.KVStoreService)
	k := Keeper{
		Environment: environment,
		cdc:         cdc,
		legacyAmino: legacyAmino,
		sk:          sk,
		authority:   authority,
		Params:      collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		ValidatorSigningInfo: collections.NewMap(
			sb,
			types.ValidatorSigningInfoKeyPrefix,
			"validator_signing_info",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorSigningInfo](cdc),
		),
		AddrPubkeyRelation: collections.NewMap(
			sb,
			types.AddrPubkeyRelationKeyPrefix,
			"addr_pubkey_relation",
			sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			codec.CollInterfaceValue[cryptotypes.PubKey](cdc),
		),
		ValidatorMissedBlockBitmap: collections.NewMap(
			sb,
			types.ValidatorMissedBlockBitmapKeyPrefix,
			"validator_missed_block_bitmap",
			collections.PairKeyCodec(sdk.LengthPrefixedBytesKey, collections.Uint64Key),
			collections.BytesValue,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/symSlash module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetPubkey(ctx context.Context, a cryptotypes.Address) (cryptotypes.PubKey, error) {
	return k.AddrPubkeyRelation.Get(ctx, a)
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k Keeper) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	err := k.sk.Jail(ctx, consAddr)
	if err != nil {
		return err
	}
	consStr, err := k.sk.ConsensusAddressCodec().BytesToString(consAddr)
	if err != nil {
		return err
	}

	if err := k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeSlash,
		event.NewAttribute(types.AttributeKeyJailed, consStr),
	); err != nil {
		return err
	}
	return nil
}
