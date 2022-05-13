package http

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
)

// RateLimiter used to define API request rate limiter
type RateLimiter interface {
	// Detect if rate limit reached on @key
	Limited(key string) bool

	// Update rate limite exclusively
	UpdateRate(float64)
}

// TODO: we should implemente a common sense rate limiter

// RequestKey is a callback used to calculate request @r's ID, we can
// use r.Method + tkn + r.URL.Path as the id of @r, if the ID is empty
// string, it's degrade into a simple rate limiter(all API's request
// are limited under the rate)
type RequestKey func(r *http.Request) string

// DefaultRequestKey used to get key of HTTP request if you don't know how
// to get the key.
func DefaultRequestKey(r *http.Request) string {
	return r.Header.Get("X-Forwarded-For") + r.RemoteAddr + r.Method + r.Proto + r.URL.String()
}

type APIRateLimiter interface {
	RequestLimited(*http.Request) bool
	// If rate limited, do anything what you want(cache the request, or do nothing)
	LimitReadchedCallback(*http.Request)
	// Update rate limite exclusively
	UpdateRate(float64)
}

// APIRateLimiterImpl is default implemented of APIRateLimiter based on tollbooth
type APIRateLimiterImpl struct {
	*limiter.Limiter
	rk RequestKey
}

func NewAPIRateLimiter(rate float64, rk RequestKey) *APIRateLimiterImpl {
	return &APIRateLimiterImpl{
		Limiter: tollbooth.NewLimiter(rate,
			&limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}).SetBurst(1),
		rk: rk,
	}
}

// RequestLimited used to limit query param key's value on all APIs,
// it means, if @key is `token', for specific token=abc123, if limit is 100/second,
// then token `abc123' can only request 100 APIs per second, no matter whiching
// API the token request.
func (rl *APIRateLimiterImpl) RequestLimited(r *http.Request) bool {
	if rl.rk == nil {
		return false // no RequestKey callback set, always pass
	}

	return rl.Limiter.LimitReached(rl.rk(r))
}

// LimitReadchedCallback do nothing, just drop the request
func (rl *APIRateLimiterImpl) LimitReadchedCallback(r *http.Request) {
	// do nothing
}

// UpdateRate update limite rate exclusively
func (rl *APIRateLimiterImpl) UpdateRate(rate float64) {
	rl.Limiter.SetMax(rate)
}
