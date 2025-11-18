package domain

type AuthStatus string

const (
	AuthDenied  AuthStatus = "denied"
	AuthGranted AuthStatus = "granted"
	AuthUnknown AuthStatus = "unknown"
)

type AuthResponse struct {
	OK AuthStatus `json:"ok"`
}
