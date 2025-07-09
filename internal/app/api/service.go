package api

import "marketflow/internal/app"

type APIService struct {
	repo app.APIRepo
}

func NewService(repo app.APIRepo) *APIService {
	return &APIService{repo: repo}
}
