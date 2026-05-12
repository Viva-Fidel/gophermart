package user

// Тело POST /api/user/register
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Тело POST /api/user/login
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
