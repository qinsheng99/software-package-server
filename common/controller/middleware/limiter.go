package middleware

import (
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	commonstl "github.com/opensourceways/software-package-server/common/controller"
)

const (
	errorTooManyRequest = "too many request"
)

type LimiterConfig struct {
	// Burst example Initialize the number of token
	Burst int `json:"burst"      required:"true"`
	// Limit number of new token per second
	Limit float64 `json:"limit"  required:"true"`
}

var l *ipRateLimiter

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		limit: r,
		burst: burst,
	}
}

func InitLimiter(cfg *LimiterConfig) {
	l = newIPRateLimiter(rate.Limit(cfg.Limit), cfg.Burst)
}

func IpRateLimiter() *ipRateLimiter {
	return l
}

type ipRateLimiter struct {
	ips   map[string]*rate.Limiter
	mu    *sync.RWMutex
	limit rate.Limit
	burst int
}

func (i *ipRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.limit, i.burst)

	i.ips[ip] = limiter

	return limiter
}

func (i *ipRateLimiter) Limiter(ip string) *rate.Limiter {
	i.mu.Lock()

	limiter, exists := i.ips[ip]
	if !exists {
		i.mu.Unlock()

		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}

func Limiter(ctx *gin.Context) {
	limit := l.Limiter(ctx.RemoteIP())

	if !limit.Allow() {
		commonstl.SendFailedResp(ctx, errorTooManyRequest, errors.New(errorTooManyRequest))

		ctx.Abort()
	} else {
		ctx.Next()
	}
}
