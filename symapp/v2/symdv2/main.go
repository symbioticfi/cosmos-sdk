package main

import (
	"fmt"
	"os"

	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/symapp/v2"
	"cosmossdk.io/symapp/v2/symdv2/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd[transaction.Tx]()
	if err := serverv2.Execute(rootCmd, clientv2helpers.EnvPrefix, symapp.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
