package responses

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Arrayed struct {
	Data any `json:"data"`
}
