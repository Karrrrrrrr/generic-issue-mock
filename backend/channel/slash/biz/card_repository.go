package biz

import (
	"context"
	"time"

	"generic-mock/enums"
	"generic-mock/model"
)

type CardCountRequest struct {
	AccountIDs  []model.ID
	Offset      int
	Limit       int
	CardNumber  *string
	Statuses    []enums.CardStatus
	IDs         []model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type CardExistByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type CardFindByAccountIDRequest struct {
	AccountID *model.ID
	ID        model.ID
}

type FindCardRequest struct {
	AccountID model.ID
	ID        model.ID
}

type CardListRequest struct {
	AccountIDs  []model.ID
	Offset      int
	Limit       int
	CardNumber  *string
	Statuses    []enums.CardStatus
	IDs         []model.ID
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

type SlashCardRepository interface {
	ExistByAccountID(context.Context, *CardExistByAccountIDRequest) (bool, error)
	FindCard(context.Context, *FindCardRequest) (*model.Card, error)
	Create(context.Context, *model.Card) error
	ExistByID(context.Context, model.ID) (bool, error)
	FindByID(context.Context, model.ID) (*model.Card, error)
	Count(context.Context, *CardCountRequest) (int64, error)
	List(context.Context, *CardListRequest) ([]*model.Card, error)
	Save(context.Context, *model.Card) error
	FindByAccountID(context.Context, *CardFindByAccountIDRequest) (*model.Card, error)
}
