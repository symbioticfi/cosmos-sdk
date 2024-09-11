package keeper

import (
	"context"
	"cosmossdk.io/math"
	stakingtypes "cosmossdk.io/x/symStaking/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// Struct to unmarshal the response from the Beacon Chain API
type Block struct {
	Data struct {
		Message struct {
			Body struct {
				ExecutionPayload struct {
					BlockHash string `json:"block_hash"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Validator struct {
	Stake    *big.Int
	ConsAddr [32]byte
}

const (
	BEACON_GENESIS_TIMESTAMP        = 1695902400
	SYMBIOTIC_SYNC_PERIOD           = 10
	SLOTS_IN_EPOCH                  = 32
	SLOT_DURATION                   = 12
	BLOCK_PATH                      = "/eth/v2/beacon/blocks/"
	GET_VALIDATOR_SET_FUNCTION_NAME = "getValidatorSet"
	GET_EPOCH_AT_TS_FUNCTION_NAME   = "getEpochAtTs"
	CONTRACT_ABI                    = `[
		{
			"type": "function",
			"name": "getEpochAtTs",
			"inputs": [
				{
					"name": "timestamp",
					"type": "uint48",
					"internalType": "uint48"
				}
			],
			"outputs": [
				{
					"name": "epoch",
					"type": "uint48",
					"internalType": "uint48"
				}
			],
			"stateMutability": "view"
		},	
		{
			"type": "function",
			"name": "getValidatorSet",
			"inputs": [
				{
					"name": "epoch",
					"type": "uint48",
					"internalType": "uint48"
				}
			],
			"outputs": [
				{
					"name": "validatorsData",
					"type": "tuple[]",
					"internalType": "struct SimpleMiddleware.ValidatorData[]",
					"components": [
						{
							"name": "stake",
							"type": "uint256",
							"internalType": "uint256"
						},
						{
							"name": "consAddr",
							"type": "bytes32",
							"internalType": "bytes32"
						}
					]
				}
			],
			"stateMutability": "view"
		}
	]`
)

func (k Keeper) SymbioticUpdateValidatorsPower(ctx context.Context) (string, error) {
	if k.networkMiddlewareAddress == "" {
		panic("middleware address is not set")
	}

	if k.HeaderService.HeaderInfo(ctx).Height%SYMBIOTIC_SYNC_PERIOD != 0 {
		return "", nil
	}

	blockHash, err := k.getFinalizedBlockHash(ctx)
	if err != nil {
		k.apiUrls.RotateBeaconUrl()
		return "", err
	}

	validators, err := k.getSymbioticValidatorSet(ctx, blockHash)
	if err != nil {
		k.apiUrls.RotateEthUrl()
		return "", err
	}

	for _, v := range validators {
		val, err := k.GetValidatorByConsAddr(ctx, v.ConsAddr[:20])
		if err != nil {
			if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
				continue
			}
			return "", err
		}

		k.SetValidatorTokens(ctx, val, math.NewIntFromBigInt(v.Stake))
	}

	return blockHash, nil
}

func (k Keeper) getFinalizedBlockHash(ctx context.Context) (string, error) {
	slot := k.getSlot(ctx)
	block, err := k.parseBlock(slot)
	for errors.Is(err, stakingtypes.ErrSymbioticNotFound) { // some slots on api may be omitted
		for i := 1; i < SLOTS_IN_EPOCH; i++ {
			block, err = k.parseBlock(slot + i)
			if err == nil {
				break
			}
			if !errors.Is(err, stakingtypes.ErrSymbioticNotFound) {
				return "", err
			}
		}
	}
	if err != nil {
		return "", err
	}

	return block.Data.Message.Body.ExecutionPayload.BlockHash, nil
}

func (k Keeper) getSymbioticValidatorSet(ctx context.Context, blockHash string) ([]Validator, error) {
	client, err := ethclient.Dial(k.apiUrls.GetEthApiUrl())
	if err != nil {
		return nil, err
	}
	defer client.Close()

	contractABI, err := abi.JSON(strings.NewReader(CONTRACT_ABI))
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(k.networkMiddlewareAddress)

	data, err := contractABI.Pack(GET_EPOCH_AT_TS_FUNCTION_NAME, big.NewInt(k.HeaderService.HeaderInfo(ctx).Time.Unix()))
	if err != nil {
		return nil, err
	}

	query := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err := client.CallContractAtHash(ctx, query, common.HexToHash(blockHash))
	if err != nil {
		return nil, err
	}

	currentEpoch := new(big.Int).SetBytes(result)

	data, err = contractABI.Pack(GET_VALIDATOR_SET_FUNCTION_NAME, currentEpoch)
	if err != nil {
		return nil, err
	}

	query = ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}
	result, err = client.CallContractAtHash(ctx, query, common.HexToHash(blockHash))
	if err != nil {
		return nil, err
	}

	var validators []Validator
	err = contractABI.UnpackIntoInterface(&validators, GET_VALIDATOR_SET_FUNCTION_NAME, result)
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func (k Keeper) getSlot(ctx context.Context) int {
	slot := (k.HeaderService.HeaderInfo(ctx).Time.Unix() - BEACON_GENESIS_TIMESTAMP) / SLOT_DURATION // get beacon slot
	slot = slot / SLOTS_IN_EPOCH * SLOTS_IN_EPOCH                                                    // first slot of epoch
	slot -= 2 * SLOTS_IN_EPOCH                                                                       // get finalized slot
	return int(slot)
}

func (k Keeper) parseBlock(slot int) (Block, error) {
	url := k.apiUrls.GetBeaconApiUrl() + BLOCK_PATH + strconv.Itoa(slot)

	var block Block
	resp, err := http.Get(url)
	if err != nil {
		return block, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return block, stakingtypes.ErrSymbioticNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return block, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return block, fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &block)
	if err != nil {
		return block, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return block, nil
}
