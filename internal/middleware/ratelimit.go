package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis   *redis.Client
	limit   int
	window  time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		limit:  limit,
		window: window,
	}
}

func (r *RateLimiter) Limit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if r.redis == nil {
				return next(c)
			}

			ip := c.RealIP()
			key := fmt.Sprintf("ratelimit:%s", ip)
			ctx := context.Background()

			current, err := r.redis.Incr(ctx, key).Result()
			if err != nil {
				return next(c)
			}

			if current == 1 {
				r.redis.Expire(ctx, key, r.window)
			}

			if current > int64(r.limit) {
				ttl, _ := r.redis.TTL(ctx, key).Result()
				c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.limit))
				c.Response().Header().Set("X-RateLimit-Remaining", "0")
				c.Response().Header().Set("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))

				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Too many requests. Please try again later.",
					},
				})
			}

			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.limit))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", r.limit-int(current)))

			return next(c)
		}
	}
}

func (r *RateLimiter) LimitByEndpoint(endpoint string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if r.redis == nil {
				return next(c)
			}

			ip := c.RealIP()
			key := fmt.Sprintf("ratelimit:%s:%s", endpoint, ip)
			ctx := context.Background()

			current, err := r.redis.Incr(ctx, key).Result()
			if err != nil {
				return next(c)
			}

			if current == 1 {
				r.redis.Expire(ctx, key, r.window)
			}

			if current > int64(r.limit) {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Too many requests. Please try again later.",
					},
				})
			}

			return next(c)
		}
	}
}
