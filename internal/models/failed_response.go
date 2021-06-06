package models

type FailedResponse struct {
	Message string `json:"message"`
}

func test() {
	FailedResponse{}.MarshalJSON()
}