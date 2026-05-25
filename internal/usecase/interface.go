package usecase

import "search_statistics/internal/domain"

type Service interface {
	GetTop(dto *GetTopDTO) *OutputDTO
	AddInStopList(dto *AddInStopListDTO) *OutputDTO
	DeleteFromStopList(dto *DeleteFromStopListDTO) *OutputDTO
}

type Repository interface {
	AddInStopList(word string) error
	DeleteFromStopList(word string) error
	GetTopCached(n int) []domain.QueryCount
	Update(newData map[string]int) 
}