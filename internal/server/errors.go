package server

import "errors"

var (
	ErrInvalidTopCexInRequest       = errors.New("invalid cex in request")
	ErrInvalidTopCexOutRequest      = errors.New("invalid cex out request")
	ErrInvalidTopNetInRequest       = errors.New("invalid net in request")
	ErrInvalidTopNetOutRequest      = errors.New("invalid net out request")
	ErrInvalidTopUserProfitRequest  = errors.New("invalid user profit request")
	ErrInvalidTopTokenProfitRequest = errors.New("invalid token profit request")
	ErrInvalidTokenInspect          = errors.New("invalid token inspect")
	ErrInvalidUserInspect           = errors.New("invalid user inspect")
	ErrInvalidDuration              = errors.New("invalid duration")

	ErrInvalidListToken     = errors.New("invalid get list token")
	ErrInvalidListUser      = errors.New("invalid get list user")
	ErrInvalidGetActivities = errors.New("invalid get activities")

	ErrInvalidGetLeaderboard  = errors.New("invalid get leaderboard")
	ErrInvalidGetUserBalances = errors.New("invalid get user balances")
	ErrInvalidGetTokenInfo    = errors.New("invalid get token info")
)
