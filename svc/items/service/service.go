package service

import (
	"context"
	"items/repo"
	itemsv1 "proto/out/items/v1"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	GetItems(ctx context.Context, categories []string) (*itemsv1.GetItemsResponse, error)
	GetItemById(ctx context.Context, id string) (*itemsv1.GetItemResponse, error)
}

type service struct {
	repo *repo.Repo
}

func NewService(pool *pgxpool.Pool) *service {
	return &service{
		repo: repo.NewRepo(pool),
	}
}

func (s *service) GetItems(ctx context.Context, categories []string) (*itemsv1.GetItemsResponse, error) {
	items, err := s.repo.GetItems(ctx, categories)
	if err != nil {
		return nil, err
	}
	resp := &itemsv1.GetItemsResponse{
		Items: make([]*itemsv1.Item, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, &itemsv1.Item{
			ItemId:      item.ID.String(),
			Name:        item.Name,
			Description: item.Description.String,
			Categories:  item.CategoryNames,
		})
	}
	return resp, nil
}

func (s *service) GetItemById(ctx context.Context, id string) (*itemsv1.GetItemResponse, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetItemById(ctx, parsed)
	if err != nil {
		return nil, err
	}
	resp := &itemsv1.GetItemResponse{
		Item: &itemsv1.Item{
			ItemId:      item.ID.String(),
			Name:        item.Name,
			Description: item.Description.String,
			Categories:  item.CategoryNames,
		},
	}
	return resp, nil
}
