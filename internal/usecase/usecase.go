package usecase

type Usecase struct {
	r Repository
}

func NewUsecase(r Repository) *Usecase{
	return &Usecase{r: r}
}

func (u *Usecase) GetTop(dto *GetTopDTO) *OutputDTO {
	data := u.r.GetTopCached(dto.NumberOfPositions)
	return &OutputDTO{Status: "ok", Msg: "", Data: data}
}

func (u *Usecase) AddInStopList(dto *AddInStopListDTO) *OutputDTO {
	err := u.r.AddInStopList(dto.Word)
	if (err != nil) {
		return &OutputDTO{Status: "error", Msg: err.Error(), Data: nil}
	}
	return &OutputDTO{Status: "ok", Msg: "", Data: nil}
}
func (u *Usecase) DeleteFromStopList(dto *DeleteFromStopListDTO) *OutputDTO {
	err := u.r.DeleteFromStopList(dto.Word)
	if (err != nil) {
		return &OutputDTO{Status: "error", Msg: err.Error(), Data: nil}
	}
	return &OutputDTO{Status: "ok", Msg: "", Data: nil}
}