package domain

type IPListStatus string

const (
	IPInBlacklist IPListStatus = "blacklist"
	IPInWhitelist IPListStatus = "whitelist"
	IPNotInList   IPListStatus = "not_in_list"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type AuthResponse struct {
	OK bool `json:"ok"`
}
