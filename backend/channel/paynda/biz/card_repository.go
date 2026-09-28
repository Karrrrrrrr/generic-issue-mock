package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardStatusExistsRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardStatusLockRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardStatusSaveRequest struct {
	AccountID              model.ID
	ID                     model.ID
	Status                 enums.CardStatus
	LastOperationRequestID string
	LastOperationType      enums.OperationType
	LastOperationStatus    enums.OperationStatus
}

type CardExistByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardExistByLastOperationRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardExistByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardFindByIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardFindByLastOperationRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type CardFindByRequestIDRequest struct {
	AccountID model.ID
	RequestID string
}

type FindCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardListRequest struct {
	AccountIDs  []model.ID
	Offset      int
	Limit       int
	IDs         []model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Statuses    []enums.CardStatus
	CardNumber  *string
}

type CardCountRequest struct {
	AccountIDs  []model.ID
	IDs         []model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Statuses    []enums.CardStatus
	CardNumber  *string
}

type CardListForFundingRequest struct {
	AccountID model.ID
	WalletIDs []model.ID
}

type PayndaCardRepository interface {
	ListForFunding(context.Context, *CardListForFundingRequest) ([]*model.Card, error)
	ExistForStatusChange(context.Context, *CardStatusExistsRequest) (bool, error)
	LockForStatusChange(context.Context, *CardStatusLockRequest) (*model.Card, error)
	SaveStatus(context.Context, *CardStatusSaveRequest) error
	Count(context.Context, *CardCountRequest) (int64, error)
	FindCard(context.Context, *FindCardRequest) (*model.Card, error)
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, *CardExistByIDRequest) (bool, error)
	ExistByRequestID(context.Context, *CardExistByRequestIDRequest) (bool, error)
	ExistByLastOperationRequestID(context.Context, *CardExistByLastOperationRequestIDRequest) (bool, error)
	FindByID(context.Context, *CardFindByIDRequest) (*model.Card, error)
	FindByRequestID(context.Context, *CardFindByRequestIDRequest) (*model.Card, error)
	FindByLastOperationRequestID(context.Context, *CardFindByLastOperationRequestIDRequest) (*model.Card, error)
	List(context.Context, *CardListRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
}
