package symGenutil

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/genesis"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/symGenutil/types"
)

var (
	_ module.HasABCIGenesis = AppModule{}

	_ appmodule.AppModule        = AppModule{}
	_ appmodulev2.GenesisDecoder = AppModule{}
)

// AppModule implements an application module for the symGenutil module.
type AppModule struct {
	cdc              codec.Codec
	accountKeeper    types.AccountKeeper
	stakingKeeper    types.StakingKeeper
	deliverTx        genesis.TxHandler
	txEncodingConfig client.TxEncodingConfig
	genTxValidator   types.MessageValidator
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	deliverTx genesis.TxHandler,
	txEncodingConfig client.TxEncodingConfig,
	genTxValidator types.MessageValidator,
) module.AppModule {
	return AppModule{
		cdc:              cdc,
		accountKeeper:    accountKeeper,
		stakingKeeper:    stakingKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
		genTxValidator:   genTxValidator,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the symGenutil module's name.
// Deprecated: kept for legacy reasons.
func (AppModule) Name() string {
	return types.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the symGenutil module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the symGenutil module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(&data, am.txEncodingConfig.TxJSONDecoder(), am.genTxValidator)
}

// InitGenesis performs genesis initialization for the symGenutil module.
func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) ([]module.ValidatorUpdate, error) {
	var genesisState types.GenesisState
	am.cdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.stakingKeeper, am.deliverTx, genesisState, am.txEncodingConfig)
}

// ExportGenesis returns the exported genesis state as raw bytes for the symGenutil module.
func (am AppModule) ExportGenesis(_ context.Context) (json.RawMessage, error) {
	return am.DefaultGenesis(), nil
}

// GenTxValidator returns the symGenutil module's genesis transaction validator.
func (am AppModule) GenTxValidator() types.MessageValidator {
	return am.genTxValidator
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) DecodeGenesisJSON(data json.RawMessage) ([]json.RawMessage, error) {
	var genesisState types.GenesisState
	if err := am.cdc.UnmarshalJSON(data, &genesisState); err != nil {
		return nil, err
	}
	return genesisState.GenTxs, nil
}
