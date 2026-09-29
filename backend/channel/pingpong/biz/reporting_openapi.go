package biz

import (
	"context"
	"time"

	pingerrors "generic-mock/channel/pingpong/errors"
	common "generic-mock/enums"
	"generic-mock/model"
	"generic-mock/pkg/types"
	sharedbiz "generic-mock/shared/biz"

	"go.uber.org/zap"
)

type AccountBalanceReportRequest struct {
	AccountID         model.ID
	Currencies        []common.Currency
	VirtualAccountIDs []model.ID
	Offset            int
	Limit             *int
}

type CardTransactionReportRequest struct {
	AccountID   model.ID
	CardID      *model.ID
	Types       []common.CardTransactionType
	Statuses    []common.CardTransactionStatus
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Offset      int
	Limit       *int
}

type AccountTransactionReportRequest struct {
	AccountID        model.ID
	VirtualAccountID *model.ID
	CardID           *model.ID
	Types            []common.CardTransactionType
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
	Offset           int
	Limit            *int
}

type TransactionReportItem struct {
	Transaction *model.CardTransaction
	Card        *model.Card
}

func (uc *PingPongOpenAPIUsecase) AccountBalanceReport(
	ctx context.Context,
	req *AccountBalanceReportRequest,
) ([]*model.VirtualAccount, int64, error) {
	if req.AccountID <= 0 || (req.Limit != nil && *req.Limit < 1) {
		return nil, 0, pingerrors.ErrInvalid
	}
	for _, id := range req.VirtualAccountIDs {
		if id <= 0 {
			return nil, 0, pingerrors.ErrInvalid
		}
	}
	for _, currency := range req.Currencies {
		if currency != common.Currency_USD {
			return nil, 0, pingerrors.ErrInvalid
		}
	}
	filters := &VirtualAccountListRequest{
		AccountIDs: []model.ID{req.AccountID},
		IDs:        req.VirtualAccountIDs,
		Offset:     req.Offset,
		Limit:      req.Limit,
	}
	items, err := uc.virtualAccountRepo.List(ctx, filters)
	if err != nil {
		zap.S().Errorw("list pingpong account balance report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	total, err := uc.virtualAccountRepo.Count(ctx, &VirtualAccountCountRequest{
		AccountIDs: []model.ID{req.AccountID},
		IDs:        req.VirtualAccountIDs,
	})
	if err != nil {
		zap.S().Errorw("count pingpong account balance report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return items, total, nil
}

func (uc *PingPongOpenAPIUsecase) CardTransactionReport(
	ctx context.Context,
	req *CardTransactionReportRequest,
) ([]TransactionReportItem, int64, error) {
	if req.AccountID <= 0 || (req.Limit != nil && *req.Limit < 1) {
		return nil, 0, pingerrors.ErrInvalid
	}
	filters := sharedbiz.CardTransactionFilters{
		Channel:     common.Channel_PingPong,
		AccountIDs:  []model.ID{req.AccountID},
		CardIDs:     types.PointerSlice(req.CardID),
		Types:       req.Types,
		Statuses:    req.Statuses,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
	}
	items, err := uc.sharedCardTransactionRepo.List(ctx, &sharedbiz.CardTransactionListRequest{
		CardTransactionFilters: filters,
		Offset:                 req.Offset,
		Limit:                  req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong card transaction report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	result, err := uc.attachReportCards(ctx, req.AccountID, items)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.sharedCardTransactionRepo.Count(ctx, &sharedbiz.CardTransactionCountRequest{
		CardTransactionFilters: filters,
	})
	if err != nil {
		zap.S().Errorw("count pingpong card transaction report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return result, total, nil
}

func (uc *PingPongOpenAPIUsecase) AccountTransactionReport(
	ctx context.Context,
	req *AccountTransactionReportRequest,
) ([]TransactionReportItem, int64, error) {
	if req.AccountID <= 0 || (req.Limit != nil && *req.Limit < 1) {
		return nil, 0, pingerrors.ErrInvalid
	}
	filters := sharedbiz.CardTransactionFilters{
		Channel:           common.Channel_PingPong,
		AccountIDs:        []model.ID{req.AccountID},
		CardIDs:           types.PointerSlice(req.CardID),
		VirtualAccountIDs: types.PointerSlice(req.VirtualAccountID),
		Types:             req.Types,
		CreatedFrom:       req.CreatedFrom,
		CreatedTo:         req.CreatedTo,
	}
	items, err := uc.sharedCardTransactionRepo.List(ctx, &sharedbiz.CardTransactionListRequest{
		CardTransactionFilters: filters,
		Offset:                 req.Offset,
		Limit:                  req.Limit,
	})
	if err != nil {
		zap.S().Errorw("list pingpong account transaction report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	result, err := uc.attachReportCards(ctx, req.AccountID, items)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.sharedCardTransactionRepo.Count(ctx, &sharedbiz.CardTransactionCountRequest{
		CardTransactionFilters: filters,
	})
	if err != nil {
		zap.S().Errorw("count pingpong account transaction report", "error", err)
		return nil, 0, pingerrors.ErrDatabase
	}
	return result, total, nil
}

func (uc *PingPongOpenAPIUsecase) attachReportCards(
	ctx context.Context,
	accountID model.ID,
	transactions []*model.CardTransaction,
) ([]TransactionReportItem, error) {
	cardIDSet := make(map[model.ID]struct{}, len(transactions))
	cardIDs := make([]model.ID, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.CardID <= 0 {
			continue
		}
		if _, exists := cardIDSet[transaction.CardID]; exists {
			continue
		}
		cardIDSet[transaction.CardID] = struct{}{}
		cardIDs = append(cardIDs, transaction.CardID)
	}
	cardByID := make(map[model.ID]*model.Card, len(cardIDs))
	if len(cardIDs) != 0 {
		cards, err := uc.sharedCardRepo.List(ctx, &sharedbiz.CardListRequest{
			CardFilters: sharedbiz.CardFilters{
				Channel:    common.Channel_PingPong,
				AccountIDs: []model.ID{accountID},
				IDs:        cardIDs,
			},
		})
		if err != nil {
			zap.S().Errorw("list pingpong report cards", "error", err)
			return nil, pingerrors.ErrDatabase
		}
		for _, card := range cards {
			cardByID[card.ID] = card
		}
	}
	result := make([]TransactionReportItem, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, TransactionReportItem{
			Transaction: transaction,
			Card:        cardByID[transaction.CardID],
		})
	}
	return result, nil
}
