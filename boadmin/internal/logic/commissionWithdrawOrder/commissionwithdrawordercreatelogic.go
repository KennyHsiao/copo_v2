package commissionWithdrawOrder

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"
	"mime/multipart"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommissionWithdrawOrderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommissionWithdrawOrderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) CommissionWithdrawOrderCreateLogic {
	return CommissionWithdrawOrderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommissionWithdrawOrderCreateLogic) CommissionWithdrawOrderCreate(req *types.CommissionWithdrawOrderCreateRequest, files []*multipart.FileHeader) (err error) {

	if len(files) > 0 {
		if req.AttachmentPath, err = utils.FileUpload(files[0], []string{}, "./public/uploads/commissionWithdrawOrder/"); err != nil {
			return err
		}
	}

	var rpcRequest transaction.WithdrawCommissionOrderRequest
	copier.Copy(&rpcRequest, &req)
	rpcRequest.CreatedBy = l.ctx.Value("account").(string)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.WithdrawCommissionOrderTransaction(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "WithdrawCommissionOrderTransaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
