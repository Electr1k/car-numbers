package provider

import "errors"

var (
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrInvalidResponse     = errors.New("invalid provider response")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrNoMoreData          = errors.New("no more data available")
	ErrParsingFailed       = errors.New("failed to parse provider data")
)
