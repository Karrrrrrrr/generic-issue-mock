package errors

import kratoserrors "github.com/go-kratos/kratos/v2/errors"

var (
	ErrInvalid      = kratoserrors.BadRequest("INVALID_PARAMETER", "参数或资源关系无效")
	ErrNotFound     = kratoserrors.NotFound("RESOURCE_NOT_FOUND", "资源不存在")
	ErrDatabase     = kratoserrors.InternalServer("DATABASE_ERROR", "数据库操作失败")
	ErrInsufficient = kratoserrors.BadRequest("INSUFFICIENT_BALANCE", "卡片或来源钱包可用余额不足")
	ErrConflict     = kratoserrors.Conflict("REQUEST_CONFLICT", "请求编号已用于不同操作")
	ErrClosed       = kratoserrors.BadRequest("CARD_CLOSED", "注销卡不能恢复或操作")
	ErrUnsupported  = kratoserrors.New(501, "NOT_IMPLEMENTED", "该 PingPong 协议尚未实现")
	ErrAppMapping   = kratoserrors.BadRequest("APP_ACCOUNT_NOT_CONFIGURED", "请配置 PINGPONG_APP_ACCOUNTS 中的应用账户映射")
)
