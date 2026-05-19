package balance

// Тело POST /api/user/balance/withdraw.
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Тело GET /api/user/withdrawals.
type ListWithdrawalsResponse []Withdrawal
