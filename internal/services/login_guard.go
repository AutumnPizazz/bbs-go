package services

import (
	"sync"
	"time"

	"bbs-go/internal/pkg/locales"

	"github.com/goburrow/cache"
)

// loginGuard protects against brute-force attacks with two layers:
//   - IP-based rate limiting (global throttle)
//   - Account-based lockout (per-user after consecutive failures)

const (
	loginIPWindow       = 60 * time.Second
	loginIPMaxAttempts  = 10
	loginLockThreshold  = 5
	loginLockDuration   = 15 * time.Minute
)

var LoginGuard = newLoginGuard()

type loginGuardService struct {
	mu       sync.Mutex
	ipCache  cache.Cache
	userCache cache.Cache
}

type loginIPState struct {
	count     int
	windowEnd int64
}

type loginUserState struct {
	failures  int
	lockedUntil int64
}

func newLoginGuard() *loginGuardService {
	return &loginGuardService{
		ipCache: cache.New(
			cache.WithMaximumSize(10000),
			cache.WithExpireAfterAccess(5 * time.Minute),
		),
		userCache: cache.New(
			cache.WithMaximumSize(5000),
			cache.WithExpireAfterAccess(30 * time.Minute),
		),
	}
}

// CheckIP returns an error if the IP has exceeded the rate limit.
func (g *loginGuardService) CheckIP(ip string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	val, ok := g.ipCache.GetIfPresent(ip)
	var state *loginIPState
	if ok {
		state = val.(*loginIPState)
	} else {
		state = &loginIPState{windowEnd: now + loginIPWindow.Milliseconds()}
	}

	if now > state.windowEnd {
		// Window expired, reset
		state.count = 0
		state.windowEnd = now + loginIPWindow.Milliseconds()
	}

	state.count++
	g.ipCache.Put(ip, state)

	if state.count > loginIPMaxAttempts {
		return ErrLoginTooMany
	}
	return nil
}

// CheckUser returns an error if the user account is locked.
func (g *loginGuardService) CheckUser(userId int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	val, ok := g.userCache.GetIfPresent(userId)
	if !ok {
		return nil
	}
	state := val.(*loginUserState)
	if state.lockedUntil > 0 && time.Now().UnixMilli() < state.lockedUntil {
		return ErrLoginAccountLocked
	}
	return nil
}

// RecordFailure increments the failure counter for a user.
// If the threshold is reached, the account is locked.
func (g *loginGuardService) RecordFailure(userId int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	val, ok := g.userCache.GetIfPresent(userId)
	var state *loginUserState
	if ok {
		state = val.(*loginUserState)
	} else {
		state = &loginUserState{}
	}

	state.failures++
	if state.failures >= loginLockThreshold {
		state.lockedUntil = time.Now().Add(loginLockDuration).UnixMilli()
	}
	g.userCache.Put(userId, state)
}

// RecordSuccess clears the failure counter for a user.
func (g *loginGuardService) RecordSuccess(userId int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.userCache.Invalidate(userId)
}

var (
	ErrLoginTooMany       = &loginGuardError{key: "auth.login_too_many"}
	ErrLoginAccountLocked = &loginGuardError{key: "auth.login_account_locked"}
)

type loginGuardError struct {
	key string
}

func (e *loginGuardError) Error() string {
	return locales.Get(e.key)
}
