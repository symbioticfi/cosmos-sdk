# Symbiotic Cosmos SDK Example

Symbiotic StubChain serves as an example network that can be built on the Symbiotic system. It is created using the Cosmos SDK with several differences.

Instead of the basic delegation module, validators use voting power as stakes in Symbiotic. To get the stakes, a validator first retrieves the last finalized block from the BeaconChain Client and then calls Network Middleware view functions at the given block height. Each node requires its own BeaconChain client. The Cosmos SDK only replicates the given voting power.

There are other minor differences including:

1. **Slashing module is changed**
    
    Only inactivity penalties (jail) are retained; others are removed.
    
2. **Government module is changed**
    
    The delegation module is removed.
    
3. **GenUtil module is changed**
    
    Delegation-related arguments are removed.
    
4. **App is changed**
    
    NFT mint, Distribution, Fee grant, and Evidence are removed.
    
5. **Required environment arguments:**
    1. Middleware address - MIDDLEWARE_ADDRESS
    2. Beacon RPC URL - BEACON_API_URLS (separated by ',')
    3. ETH RPC URL - ETH_API_URLS (separated by ',')
    4. Debug flag (for unfinalized canonical block) - DEBUG [Optional]

## Modules
- /x/symStaking <- x/staking
- /x/symSlash <- x/slashing
- /x/symGov <- x/gov
- /x/symGenutil <- x/genutil

## Build
```bash
make build-sym
```

## Run
See [`symapp`](symapp/README.md) directory