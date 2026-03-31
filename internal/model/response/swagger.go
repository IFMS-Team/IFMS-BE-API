package response

type APIResponse struct {
	Status  int         `json:"status" example:"200"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Invalid parameters"`
	Error   string `json:"error,omitempty" example:"field validation failed"`
}

type PaginatedResponse struct {
	Status     int         `json:"status" example:"200"`
	Message    string      `json:"message" example:"Success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page  int   `json:"page" example:"1"`
	Size  int   `json:"size" example:"10"`
	Total int64 `json:"total" example:"100"`
}

type MessageResponse struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"Success"`
}
