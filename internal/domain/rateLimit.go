package domain

import "time"

type RateLimitRule struct {
	Action string // "create_post", "create_comment"
	Limit  int
	Window time.Duration // e.g. 1 * time.Minute
}

type RateLimitResult struct {
	IsAllowed  bool
	Remaining  int
	RetryAfter time.Duration
}

var (
	RuleCreatePost    = RateLimitRule{Action: "create_post", Limit: 3, Window: 1 * time.Minute}
	RuleCreateComment = RateLimitRule{Action: "create_comment", Limit: 5, Window: 1 * time.Minute}
)
