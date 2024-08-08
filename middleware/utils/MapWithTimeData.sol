// SPDX-License-Identifier: MIT
pragma solidity ^0.8.25;

import {Checkpoints} from "@openzeppelin/contracts/utils/structs/Checkpoints.sol";
import {Time} from "@openzeppelin/contracts/utils/types/Time.sol";
import {EnumerableMap} from "@openzeppelin/contracts/utils/structs/EnumerableMap.sol";

library MapWithTimeData {
    using EnumerableMap for EnumerableMap.AddressToUintMap;

    uint256 private constant ENABLED_TIME_MASK = 0xFFFFFFFFFFFFFFFFFFFFFFFF;
    uint256 private constant DISABLED_TIME_MASK = 0xFFFFFFFFFFFFFFFFFFFFFFFF << 48;

    function add(EnumerableMap.AddressToUintMap storage self, address addr) internal {
        self.set(addr, uint256(0));
    }

    function disable(EnumerableMap.AddressToUintMap storage self, address addr) internal {
        uint256 value = self.get(addr);
        value |= uint256(Time.timestamp()) << 48;
        self.set(addr, value);
    }

    function enable(EnumerableMap.AddressToUintMap storage self, address addr) internal {
        uint256 value = self.get(addr);
        value |= uint256(Time.timestamp());
        value &= ~DISABLED_TIME_MASK;
        self.set(addr, value);
    }

    function atWithTimes(EnumerableMap.AddressToUintMap storage self, uint256 idx)
        internal
        view
        returns (address key, uint48 enabledTime, uint48 disabledTime)
    {
        uint256 value = 0;
        (key, value) = self.at(idx);
        enabledTime = uint48(value & ENABLED_TIME_MASK);
        disabledTime = uint48((value & DISABLED_TIME_MASK) >> 48);
    }

    function getTimes(EnumerableMap.AddressToUintMap storage self, address addr)
        internal
        view
        returns (uint48 enabledTime, uint48 disabledTime)
    {
        uint256 value = self.get(addr);
        enabledTime = uint48(value & ENABLED_TIME_MASK);
        disabledTime = uint48((value & DISABLED_TIME_MASK) >> 48);
    }
}
