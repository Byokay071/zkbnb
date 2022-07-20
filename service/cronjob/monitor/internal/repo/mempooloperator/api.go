package mempooloperator

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
)

type Model interface {
	CreateMempoolTxs(pendingNewMempoolTxs []*table.MempoolTx) (err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: `mempool_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
