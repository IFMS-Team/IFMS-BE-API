package model

type SuccessResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type PaginationRequest struct {
	Page  int32 `form:"page" binding:"omitempty,min=1"`
	Limit int32 `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (p *PaginationRequest) GetOffset() int32 {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
	return (p.Page - 1) * p.Limit
}
