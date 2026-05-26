package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client  *ethclient.Client
	chainID *big.Int
}

func NewClient(rpcURL string, chainID int64) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial blockchain failed: %w", err)
	}

	return &Client{
		client:  client,
		chainID: big.NewInt(chainID),
	}, nil
}

func (c *Client) Close() {
	c.client.Close()
}

// GetBalance 查询指定地址的 ETH 余额
func (c *Client) GetBalance(ctx context.Context, address string) (string, error) {
	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return "", fmt.Errorf("get balance failed: %w", err)
	}

	// 从 wei 转为 ETH（1 ETH = 10^18 wei）
	ethValue := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		big.NewFloat(1e18),
	)
	return ethValue.String(), nil
}

// GetBlockNumber 查询最新区块号
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	return c.client.BlockNumber(ctx)
}
