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

type CardOperationUpdateRequest struct {
	AccountID model.ID
	ID        model.ID
	RequestID string
	Type      enums.OperationType
	Status    enums.OperationStatus
}

type CardExistCardByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardExistCardByLastOperationRequestIDForAccountRequest struct {
	AccountID model.ID
	RequestID string
}

type CardExistCardByRequestIDForAccountRequest struct {
	AccountID model.ID
	RequestID string
}

type FindCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardFindCardByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardFindCardByLastOperationRequestIDForAccountRequest struct {
	AccountID model.ID
	RequestID string
}

type CardFindCardByRequestIDForAccountRequest struct {
	AccountID model.ID
	RequestID string
}

type CardListCardsRequest struct {
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

type CardRepository interface {
	ListForFunding(context.Context, *CardListForFundingRequest) ([]*model.Card, error)
	UpdateOperation(context.Context, *CardOperationUpdateRequest) error
	ExistForStatusChange(context.Context, *CardStatusExistsRequest) (bool, error)
	LockForStatusChange(context.Context, *CardStatusLockRequest) (*model.Card, error)
	SaveStatus(context.Context, *CardStatusSaveRequest) error
	Count(context.Context, *CardCountRequest) (int64, error)
	FindCard(context.Context, *FindCardRequest) (*model.Card, error)
	CreateCard(context.Context, *model.Card) error
	ExistCardByID(context.Context, model.ID) (bool, error)
	FindCardByID(context.Context, model.ID) (*model.Card, error)
	ExistCardByRequestID(context.Context, string) (bool, error)
	FindByRequestID(context.Context, string) (*model.Card, error)
	ExistCardByLastOperationRequestID(context.Context, string) (bool, error)
	FindByLastOperationRequestID(context.Context, string) (*model.Card, error)
	ExistCardByRequestIDForAccount(context.Context, *CardExistCardByRequestIDForAccountRequest) (bool, error)
	FindCardByRequestIDForAccount(context.Context, *CardFindCardByRequestIDForAccountRequest) (*model.Card, error)
	ExistCardByLastOperationRequestIDForAccount(context.Context, *CardExistCardByLastOperationRequestIDForAccountRequest) (bool, error)
	FindCardByLastOperationRequestIDForAccount(context.Context, *CardFindCardByLastOperationRequestIDForAccountRequest) (*model.Card, error)
	ListCards(context.Context, *CardListCardsRequest) ([]*model.Card, error)
	SaveCard(context.Context, *model.Card) error
	ExistCardByAccountID(context.Context, *CardExistCardByAccountIDRequest) (bool, error)
	FindCardByAccountID(context.Context, *CardFindCardByAccountIDRequest) (*model.Card, error)
}
