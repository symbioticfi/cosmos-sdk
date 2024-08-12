// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

import {EnumerableMap} from "@openzeppelin/contracts/utils/structs/EnumerableMap.sol";

import {MapWithTimeData} from "src/libraries/MapWithTimeData.sol";

contract MapWithTimeDataContract {
    using EnumerableMap for EnumerableMap.AddressToUintMap;
    using MapWithTimeData for EnumerableMap.AddressToUintMap;

    EnumerableMap.AddressToUintMap internal elements;

    function add(address addr) public {
        elements.add(addr);
    }

    function disable(address addr) public {
        elements.disable(addr);
    }

    function enable(address addr) public {
        elements.enable(addr);
    }

    function atWithTimes(uint256 idx) public view returns (address key, uint48 enabledTime, uint48 disabledTime) {
        return elements.atWithTimes(idx);
    }

    function getTimes(address addr) public view returns (uint48 enabledTime, uint48 disabledTime) {
        return elements.getTimes(addr);
    }

    function wasActiveAt(address addr, uint48 timestamp) public view returns (bool) {
        (uint48 enabledTime, uint48 disabledTime) = getTimes(addr);

        return enabledTime != 0 && enabledTime <= timestamp && (disabledTime == 0 || disabledTime >= timestamp);
    }
}
