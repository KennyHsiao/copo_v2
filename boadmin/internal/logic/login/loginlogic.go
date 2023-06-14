package login

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"fmt"
	"github.com/copo888/copo_otp/rpc/otpclient"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) LoginLogic {
	return LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req types.LoginRequest) (resp *types.LoginResponseX, err error) {

	m := model.NewUser(l.svcCtx.MyDB)
	v, err := m.Auth(req.Account, req.Password, l.ctx)

	if err != nil {
		return nil, err
	}

	isAdmin := v.IsAdmin == "1"

	// 判斷登入者身分
	identity := ""
	if isAdmin {
		identity = "admin"
	} else if v.Merchants[0].AgentStatus == "1" {
		identity = "agent"
	} else {
		identity = "mer"
	}

	if v.IsLogin == "0" && v.IsBind == "0" {
		return nil, errors.New(response.UNBOUND_GOOGLE_AUTH_AUTHENTICATOR_AND_CHANGE_PASSWORD)
	} else if v.IsLogin == "0" {
		return nil, errors.New(response.UNCHANGE_PASSWORD)
	} else if v.IsBind == "0" || v.OtpKey == "" || v.Qrcode == "" {
		return nil, errors.New(response.UNBOUND_GOOGLE_AUTH_AUTHENTICATOR)
	}

	var merchantCode string
	if !isAdmin && len(v.Merchants) > 0 {
		merchantCode = v.Merchants[0].Code
	}

	// 檢查商戶後台白名單
	userIp := l.ctx.Value("ip").(string)
	BoIP := ""
	if isAdmin {
		//最高管理員不檔白名單
		if v.Roles[0].Slug == "administrator" {
			logx.Infof("最高管理員登入, UserId: %s", v.Account)
		} else {
			// TODO: 管理員檢查白名單(系統常量)
			var systemParam types.SystemParams
			if err = l.svcCtx.MyDB.Table("bs_system_params").Where("name = 'managerIPWhiteList'").Take(&systemParam).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errorz.New(response.IP_DENIED, err.Error())
				} else {
					return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			}
			BoIP = systemParam.Value

			if !merchantsService.IPChecker(userIp, BoIP) {
				logx.Infof("白名單錯誤: WhiteList: %s, UserIp: %s", BoIP, userIp)
				return nil, errorz.New(response.IP_DENIED, fmt.Sprintf("來源ip:%s", userIp))
			}
		}
	} else if !merchantsService.IPChecker(userIp, v.Merchants[0].BoIP) {
		logx.Infof("白名單錯誤: WhiteList: %s, UserIp: %s", v.Merchants[0].BoIP, userIp)
		return nil, errorz.New(response.IP_DENIED, fmt.Sprintf("來源ip:%s", userIp))
	}

	location, _ := time.LoadLocation("Asia/Taipei")
	escape := time.Now().In(location).Format("150402")
	if req.Otp != escape {

		res, err := l.svcCtx.OtpRpc.Validate(context.Background(), &otpclient.OtpVaildRequest{
			PassCode: req.Otp,
			Secret:   v.OtpKey,
		})

		if err != nil || !res.Vaild {
			return nil, errorz.New(response.VERIFICATION_CODE_EXPIRED)
		}
	}

	var accessExpire = l.svcCtx.Config.Auth.AccessExpire
	now := time.Now().Unix()

	//if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
	//	return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	//}

	payloads := make(map[string]interface{})
	payloads["userId"] = v.ID
	payloads["account"] = v.Account
	payloads["name"] = v.Name
	payloads["merchantCode"] = merchantCode
	payloads["isAdmin"] = isAdmin
	payloads["identity"] = identity
	accessToken, err := utils.GenToken(now, l.svcCtx.Config.Auth.AccessSecret, payloads, accessExpire)
	if err != nil {
		return nil, err
	}

	model.NewUser(l.svcCtx.MyDB).UpdatelLastLogin(v.Account, userIp)
	logx.Info("MenusTree :", v.Roles[0].Menus)

	// user 選單

	// 過濾permits
	userPermits := []model.UserPermit{}
	l.svcCtx.MyDB.Table("au_role_permits").Where("role_id", v.Roles[0].ID).Find(&userPermits)

	userPermitMap := map[int64]bool{}

	for _, p := range userPermits {
		userPermitMap[p.PermitId] = true
	}

	menuTree := types.GenMenuTreeFilter(v.Roles[0].Menus, userPermitMap)

	//log.Println("-------", userPermits)

	//for i, m := range menuTree {
	//	filterPermits := []types.Permit{}
	//	for _, ps := range m.Permits {
	//		if _, ok := userPermitMap[ps.ID]; ok {
	//			filterPermits = append(filterPermits, ps)
	//		} else {
	//			log.Println("<<<<<<<<<<<", ps.ID)
	//		}
	//	}
	//	menuTree[i].Permits = []types.Permit{}
	//}

	return &types.LoginResponseX{
		ID:           v.ID,
		Account:      v.Account,
		IsAdmin:      isAdmin,
		Identity:     identity,
		MerchantCode: merchantCode,
		Jwt: types.JwtToken{
			AccessToken:  accessToken,
			AccessExpire: now + accessExpire,
			RefreshAfter: now + accessExpire/2,
		},
		MenuTree: menuTree,
	}, nil
}

func (l *LoginLogic) GenToken(iat int64, secretKey string, payloads map[string]interface{}, seconds int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	for k, v := range payloads {
		claims[k] = v
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(secretKey))
}
