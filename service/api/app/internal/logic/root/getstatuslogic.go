package root

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatusLogic {
	return &GetStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func packServerVersion(CodeVersion string, GitCommitHash string) string {
	return fmt.Sprintf("%s:%s ", CodeVersion, GitCommitHash)

}

func (l *GetStatusLogic) GetStatus(req *types.ReqGetStatus) (resp *types.RespGetStatus, err error) {
	return &types.RespGetStatus{
		Status:        200,
		NetworkId:     1,
		ServerVersion: packServerVersion(l.svcCtx.CodeVersion, l.svcCtx.GitCommitHash),
	}, nil
}
