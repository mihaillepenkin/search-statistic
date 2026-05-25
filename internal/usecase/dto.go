package usecase

import "search_statistics/internal/domain"

type GetTopDTO struct {
	NumberOfPositions int `json:"number_of_positions"`
}

type AddInStopListDTO struct {
	Word string `json:"word"`
}

type DeleteFromStopListDTO struct {
	Word string `json:"word"`
}

type OutputDTO struct {
	Status string              `json:"status"`
	Msg    string              `json:"msg,omitempty"`
	Data   []domain.QueryCount "json:`data`"
}