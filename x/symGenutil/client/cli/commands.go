package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/symGenutil"
	symGenutiltypes "github.com/cosmos/cosmos-sdk/x/symGenutil/types"
)

// TODO(serverv2): remove app exporter notion that is server v1 specific

type genesisMM interface {
	DefaultGenesis() map[string]json.RawMessage
	ValidateGenesis(genesisData map[string]json.RawMessage) error
}

// Commands adds core sdk's sub-commands into genesis command.
func Commands(symGenutilModule symGenutil.AppModule, genMM genesisMM, appExport servertypes.AppExporter) *cobra.Command {
	return CommandsWithCustomMigrationMap(symGenutilModule, genMM, appExport, MigrationMap)
}

// CommandsWithCustomMigrationMap adds core sdk's sub-commands into genesis command with custom migration map.
// This custom migration map can be used by the application to add its own migration map.
func CommandsWithCustomMigrationMap(symGenutilModule symGenutil.AppModule, genMM genesisMM, appExport servertypes.AppExporter, migrationMap symGenutiltypes.MigrationMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Application's genesis-related subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		GenTxCmd(genMM, banktypes.GenesisBalancesIterator{}),
		MigrateGenesisCmd(migrationMap),
		CollectGenTxsCmd(symGenutilModule.GenTxValidator()),
		ValidateGenesisCmd(genMM),
		AddGenesisAccountCmd(),
		ExportCmd(appExport),
	)

	return cmd
}
