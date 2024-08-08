// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {SimpleMiddleware} from "src/SimpleMiddleware.sol";
import {IRegistry} from "@symbiotic/interfaces/common/IRegistry.sol";
import {INetworkRegistry} from "@symbiotic/interfaces/INetworkRegistry.sol";
import {IOperatorRegistry} from "@symbiotic/interfaces/IOperatorRegistry.sol";
import {IOptInService} from "@symbiotic/interfaces/service/IOptInService.sol";
import {IVault} from "@symbiotic/interfaces/vault/IVault.sol";
import {IBaseDelegator} from "@symbiotic/interfaces/delegator/IBaseDelegator.sol";

contract NetworkSetup is Script {
    function run(
        address networkRegistry,
        address[] memory vaults,
        uint256 subnetworksCnt,
        uint256[][] calldata networkLimits
    ) external {
        require(vaults.length == networkLimits.length, "inconsistent length");
        vm.startBroadcast();
        INetworkRegistry(networkRegistry).registerNetwork();
        for (uint256 i = 0; i < vaults.length; ++i) {
            require(subnetworksCnt == networkLimits[i].length, "inconsistent length");
            address delegator = IVault(vaults[i]).delegator();
            for (uint96 j = 0; j < subnetworksCnt; ++j) {
                IBaseDelegator(delegator).setMaxNetworkLimit(j, networkLimits[i][j]);
            }
        }
        vm.stopBroadcast();
    }
}
