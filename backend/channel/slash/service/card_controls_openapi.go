package service

import (
	"context"
	"encoding/json"

	"generic-mock/channel/slash/biz"
)

type OpenAPISpendingConstraint struct {
	MerchantCategoryRule     *json.RawMessage `json:"merchantCategoryRule,omitempty"`     // Invalid: spending rules are not persisted.
	MerchantRule             *json.RawMessage `json:"merchantRule,omitempty"`             // Invalid: spending rules are not persisted.
	SpendingRule             *json.RawMessage `json:"spendingRule,omitempty"`             // Invalid: spending rules are not persisted.
	CountryRule              *json.RawMessage `json:"countryRule,omitempty"`              // Invalid: spending rules are not persisted.
	MerchantCategoryCodeRule *json.RawMessage `json:"merchantCategoryCodeRule,omitempty"` // Invalid: spending rules are not persisted.
}

type OpenAPICardSpendingRequest struct {
	OpenAPIIDRequest
	OpenAPISpendingConstraint
}

type OpenAPICardModifierRequest struct {
	OpenAPIIDRequest
	Name  string `json:"name" binding:"required"`  // Invalid: card modifiers are not persisted.
	Value *bool  `json:"value" binding:"required"` // Invalid: card modifiers are not persisted.
}

type OpenAPIUtilization struct {
	NextResetDate    string        `json:"nextResetDate"` // Invalid: limit resets are not simulated.
	Spend            OpenAPIAmount `json:"spend"`
	AvailableBalance OpenAPIAmount `json:"availableBalance"`
}

type OpenAPICardModifier struct {
	Name  string `json:"name"`
	Value bool   `json:"value"`
}

type OpenAPICardModifiers struct {
	Modifiers []OpenAPICardModifier `json:"modifiers"`
}

func (service *SlashOpenAPIService) GetCardUtilization(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPIUtilization, error) {
	if _, err := service.GetCard(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPIUtilization{}, nil
}

func (service *SlashOpenAPIService) SetCardSpendingConstraint(ctx context.Context, req *OpenAPICardSpendingRequest) (*OpenAPISpendingConstraint, error) {
	if _, err := service.GetCard(ctx, &req.OpenAPIIDRequest); err != nil {
		return nil, err
	}
	return &OpenAPISpendingConstraint{}, nil
}

func (service *SlashOpenAPIService) GetCardModifiers(ctx context.Context, req *OpenAPIIDRequest) (*OpenAPICardModifiers, error) {
	if _, err := service.GetCard(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPICardModifiers{Modifiers: []OpenAPICardModifier{}}, nil
}

func (service *SlashOpenAPIService) SetCardModifier(ctx context.Context, req *OpenAPICardModifierRequest) (*struct{}, error) {
	_, err := service.GetCard(ctx, &req.OpenAPIIDRequest)
	return &struct{}{}, err
}

type OpenAPICardGroupRequest struct {
	OpenAPIAccountPathRequest
	OpenAPISpendingConstraint
	SpendingConstraint *OpenAPISpendingConstraint `json:"spendingConstraint"` // Invalid: group controls are not persisted.
	Name               *string                    `json:"name"`               // Invalid: group configuration is a non-persistent account view.
	VirtualAccountID   *string                    `json:"virtualAccountId"`
}

type OpenAPICardGroup struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	VirtualAccountID   string                    `json:"virtualAccountId"`
	SpendingConstraint OpenAPISpendingConstraint `json:"spendingConstraint"`
	Cards              []string                  `json:"cards"`
}

func (service *SlashOpenAPIService) CardGroup(ctx context.Context, req *OpenAPICardGroupRequest) (*OpenAPICardGroup, error) {
	account, err := service.protocolAccount(ctx, &req.OpenAPIAccountPathRequest)
	if err != nil {
		return nil, err
	}
	result := &OpenAPICardGroup{
		ID:    slashIDString(account.ID),
		Name:  account.Name,
		Cards: []string{},
	}
	if req.Name != nil {
		result.Name = *req.Name
	}
	if req.VirtualAccountID != nil {
		id, err := slashID(*req.VirtualAccountID)
		if err != nil {
			return nil, err
		}
		if _, err := service.usecase.GetVirtualAccount(ctx, &biz.ResourceRequest{
			AccountID: &account.ID,
			ID:        id,
		}); err != nil {
			return nil, err
		}
		result.VirtualAccountID = slashIDString(id)
	}
	return result, nil
}

func (service *SlashOpenAPIService) ListCardGroups(ctx context.Context, req *OpenAPICardGroupRequest) (*OpenAPIItems[*OpenAPICardGroup], error) {
	group, err := service.CardGroup(ctx, req)
	if err != nil {
		return nil, err
	}
	return &OpenAPIItems[*OpenAPICardGroup]{
		Items:    []*OpenAPICardGroup{group},
		Metadata: OpenAPIMetadata{Count: 1},
	}, nil
}

func (service *SlashOpenAPIService) SetGroupSpendingConstraint(ctx context.Context, req *OpenAPICardGroupRequest) (*OpenAPISpendingConstraint, error) {
	if _, err := service.CardGroup(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPISpendingConstraint{}, nil
}

func (service *SlashOpenAPIService) GetGroupUtilization(ctx context.Context, req *OpenAPICardGroupRequest) (*OpenAPIUtilization, error) {
	if _, err := service.CardGroup(ctx, req); err != nil {
		return nil, err
	}
	return &OpenAPIUtilization{}, nil
}
