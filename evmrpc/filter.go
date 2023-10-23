package evmrpc

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sei-protocol/sei-chain/x/evm/keeper"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type filter struct {
	fromBlock rpc.BlockNumber
	toBlock   rpc.BlockNumber
	addresses []common.Address
	topics    []common.Hash
	// todo: expiration
}

type FilterAPI struct {
	tmClient     rpcclient.Client
	keeper       *keeper.Keeper
	ctxProvider  func(int64) sdk.Context
	nextFilterId uint64
	filters      map[uint64]filter
}

func NewFilterAPI(tmClient rpcclient.Client, k *keeper.Keeper, ctxProvider func(int64) sdk.Context) *FilterAPI {
	filters := make(map[uint64]filter)
	return &FilterAPI{tmClient: tmClient, keeper: k, ctxProvider: ctxProvider, nextFilterId: 1, filters: filters}
}

func (a *FilterAPI) NewFilter(
	ctx context.Context,
	fromBlock rpc.BlockNumber,
	toBlock rpc.BlockNumber,
	addresses []common.Address,
	topics []string,
) (*uint64, error) {
	fromBlockPtr, err := getBlockNumber(ctx, a.tmClient, fromBlock)
	if err != nil {
		return nil, err
	}
	toBlockPtr, err := getBlockNumber(ctx, a.tmClient, toBlock)
	if err != nil {
		return nil, err
	}
	if fromBlockPtr == nil && toBlockPtr != nil {
		return nil, errors.New("from block is after to block")
	}
	if toBlockPtr != nil {
		if *fromBlockPtr > *toBlockPtr {
			return nil, errors.New("from block is after to block")
		}
	}
	var topicsRes []common.Hash
	if topics == nil {
		topicsRes = make([]common.Hash, 0)
	} else {
		for _, topic := range topics {
			topicsRes = append(topicsRes, common.HexToHash(topic))
		}
	}
	curFilterId := a.nextFilterId
	a.nextFilterId++
	f := filter{
		fromBlock: fromBlock,
		toBlock:   toBlock,
		addresses: addresses,
		topics:    topicsRes,
	}
	a.filters[curFilterId] = f
	return &curFilterId, nil
}

func (a *FilterAPI) GetFilterChanges(
	ctx context.Context,
	filterId uint64,
) ([]common.Hash, error) {
	return nil, nil
}

func (a *FilterAPI) GetFilterLogs(
	ctx context.Context,
	filterId uint64,
) ([]*ethtypes.Log, error) {
	logs, err := a.keeper.GetLogs(a.ctxProvider(LatestCtxHeight))
	if err != nil {
		return nil, err
	}
	filter, ok := a.filters[filterId]
	if !ok {
		return nil, errors.New("filter id not found")
	}
	filterLogs := make([]*ethtypes.Log, 0)
	for _, log := range logs {
		if uint64(filter.fromBlock.Int64()) <= log.BlockNumber && uint64(filter.toBlock.Int64()) >= log.BlockNumber {
			filterLogs = append(filterLogs, log)
		}
	}
	return filterLogs, nil
}

func (a *FilterAPI) GetLogs(
	ctx context.Context,
	blockHash common.Hash,
	fromBlock rpc.BlockNumber,
	toBlock rpc.BlockNumber,
	topics []common.Hash,
) ([]common.Hash, error) {
	// need to reach into get log!
	return nil, nil
}

func (a *FilterAPI) UninstallFilter(
	ctx context.Context,
	filterId uint64,
) (bool, error) {
	_, found := a.filters[filterId]
	if !found {
		return false, nil
	}
	delete(a.filters, filterId)
	return true, nil
}
