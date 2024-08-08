// SPDX-License-Identifier: UNLICENSED
pragma solidity 0.8.25;

import {Script} from "forge-std/Script.sol";
import {SimpleMiddleware} from "src/SimpleMiddleware.sol";

contract Setup is Script {
    function run(
        address network,
        address owner,
        uint48 epochDuration,
        address[] memory vaults,
        address[] memory operators,
        bytes32[] memory keys,
        address operatorRegistry,
        address vaultRegistry,
        address operatorNetworkOptIn
    ) external {
        require(operators.length == keys.length, "inconsistent length");
        vm.startBroadcast();

        uint48 minSlashingWindow = epochDuration; // we dont use this

        SimpleMiddleware middleware = new SimpleMiddleware(
            network, operatorRegistry, vaultRegistry, operatorNetworkOptIn, owner, epochDuration, minSlashingWindow
        );

        for (uint256 i = 0; i < vaults.length; ++i) {
            middleware.registerVault(vaults[i]);
        }

        for (uint256 i = 0; i < operators.length; ++i) {
            middleware.registerOperator(operators[i], keys[i]);
        }

        vm.stopBroadcast();
    }
}
