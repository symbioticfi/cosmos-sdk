package symStaking

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	_ "cosmossdk.io/api/cosmos/crypto/ed25519" // register to that it shows up in protoregistry.GlobalTypes
	stakingv1beta "cosmossdk.io/api/cosmos/symStaking/v1beta1"

	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: stakingv1beta.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Validators",
					Short:     "Query for all validators",
					Long:      "Query details about all validators on a network.",
				},
				{
					RpcMethod: "Validator",
					Use:       "validator [validator-addr]",
					Short:     "Query a validator",
					Long:      "Query details about an individual validator.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "HistoricalInfo",
					Use:       "historical-info [height]",
					Short:     "Query historical info at given height",
					Example:   fmt.Sprintf("$ %s query staking historical-info 5", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "height"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current staking parameters information",
					Long:      "Query values set as staking parameters.",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: stakingv1beta.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "UpdateParams",
					Use:            "update-params-proposal [params]",
					Short:          "Submit a proposal to update staking module params. Note: the entire params must be provided.",
					Long:           fmt.Sprintf("Submit a proposal to update staking module params. Note: the entire params must be provided.\n See the fields to fill in by running `%s query staking params --output json`", version.AppName),
					Example:        fmt.Sprintf(`%s tx staking update-params-proposal '{ "unbonding_time": "504h0m0s", ... }'`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "params"}},
					GovProposal:    true,
				},
			},
			EnhanceCustomCommand: true,
		},
	}
}
