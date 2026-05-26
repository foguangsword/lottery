package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"backend/internal/blockchain"
	"backend/pkg/logger"
)

type Scheduler struct {
	cron   *cron.Cron
	client *blockchain.Client
}

func New(client *blockchain.Client) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		client: client,
	}
}

func (s *Scheduler) Start() {
	// 示例：每 10 秒查询一次最新区块号
	s.cron.AddFunc("*/10 * * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		num, err := s.client.GetBlockNumber(ctx)
		if err != nil {
			logger.Log.Error("query block number failed", zap.Error(err))
			return
		}
		logger.Log.Info("latest block number", zap.Uint64("number", num))
	})

	// 示例：每天 0 点执行的任务
	// s.cron.AddFunc("0 0 0 * * *", func() { ... })

	s.cron.Start()
	logger.Log.Info("scheduler started")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Log.Info("scheduler stopped")
}
