package domain

type ResetBucketsRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

type ResetBucketsResponse struct {
	Reset bool `json:"reset"`
}
