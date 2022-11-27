package responses

type LogResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type LoginResponse struct {
	Message string `json:"message"`
}
