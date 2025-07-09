package api

import "marketflow/internal/app"

type APIService struct {
	repo app.AggregatedRepo
}

func NewService(repo app.AggregatedRepo) *APIService {
	return &APIService{repo: repo}
}
