// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {Test, console2} from "forge-std/Test.sol";

import {MapWithTimeData} from "src/libraries/MapWithTimeData.sol";
import {MapWithTimeDataContract} from "./mocks/MapWithTimeDataContract.sol";

contract DefaultOperatorRewardsTest is Test {
    address owner;
    address alice;
    uint256 alicePrivateKey;
    address bob;
    uint256 bobPrivateKey;

    MapWithTimeDataContract mapWithTimeDataContract;

    function setUp() public {
        owner = address(this);
        (alice, alicePrivateKey) = makeAddrAndKey("alice");
        (bob, bobPrivateKey) = makeAddrAndKey("bob");

        mapWithTimeDataContract = new MapWithTimeDataContract();
    }

    function test_All() public {
        uint256 blockTimestamp = block.timestamp * block.timestamp / block.timestamp * block.timestamp / block.timestamp;
        blockTimestamp = blockTimestamp + 1_720_700_948;
        vm.warp(blockTimestamp);

        mapWithTimeDataContract.add(alice);

        vm.expectRevert(MapWithTimeData.AlreadyAdded.selector);
        mapWithTimeDataContract.add(alice);

        (address key, uint48 enabledTime, uint48 disabledTime) = mapWithTimeDataContract.atWithTimes(0);
        assertEq(key, alice);
        assertEq(enabledTime, 0);
        assertEq(disabledTime, 0);

        (enabledTime, disabledTime) = mapWithTimeDataContract.getTimes(alice);
        assertEq(enabledTime, 0);
        assertEq(disabledTime, 0);

        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp)), false);

        mapWithTimeDataContract.enable(alice);

        vm.expectRevert(MapWithTimeData.AlreadyEnabled.selector);
        mapWithTimeDataContract.enable(alice);

        (key, enabledTime, disabledTime) = mapWithTimeDataContract.atWithTimes(0);
        assertEq(key, alice);
        assertEq(enabledTime, blockTimestamp);
        assertEq(disabledTime, 0);

        (enabledTime, disabledTime) = mapWithTimeDataContract.getTimes(alice);
        assertEq(enabledTime, blockTimestamp);
        assertEq(disabledTime, 0);

        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp)), true);

        blockTimestamp = blockTimestamp + 1;
        vm.warp(blockTimestamp);

        mapWithTimeDataContract.disable(alice);

        vm.expectRevert(MapWithTimeData.NotEnabled.selector);
        mapWithTimeDataContract.disable(alice);

        (key, enabledTime, disabledTime) = mapWithTimeDataContract.atWithTimes(0);
        assertEq(key, alice);
        assertEq(enabledTime, blockTimestamp - 1);
        assertEq(disabledTime, blockTimestamp);

        (enabledTime, disabledTime) = mapWithTimeDataContract.getTimes(alice);
        assertEq(enabledTime, blockTimestamp - 1);
        assertEq(disabledTime, blockTimestamp);

        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp - 2)), false);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp - 1)), true);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp)), true);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp + 1)), false);

        blockTimestamp = blockTimestamp + 1;
        vm.warp(blockTimestamp);

        mapWithTimeDataContract.enable(alice);

        (key, enabledTime, disabledTime) = mapWithTimeDataContract.atWithTimes(0);
        assertEq(key, alice);
        assertEq(enabledTime, blockTimestamp);
        assertEq(disabledTime, 0);

        (enabledTime, disabledTime) = mapWithTimeDataContract.getTimes(alice);
        assertEq(enabledTime, blockTimestamp);
        assertEq(disabledTime, 0);

        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp - 2)), false);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp - 1)), false);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp)), true);
        assertEq(mapWithTimeDataContract.wasActiveAt(alice, uint48(blockTimestamp + 1)), true);
    }
}
