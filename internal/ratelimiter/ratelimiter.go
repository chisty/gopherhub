package ratelimiter

import "time"

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTimeFrame int
	TimeFrame           time.Duration
	Enabled             bool
}

// Should return http.StatusCode 429 (Too many requests) if the rate limit is exceeded
// Should include a Retry-After header with the time until the next request is allowed
