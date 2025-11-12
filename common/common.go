package common

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"braces.dev/errtrace"
)

const (
	IP_CHECK_HOST = "api.ipify.org"
)

func RunWithTimeout[T any](f func() *T, deadline time.Time) (*T, error) {
	ctx, stop := context.WithDeadline(context.Background(), deadline)
	start := time.Now()
	defer stop()

	var done chan *T

	func(ctx context.Context) {
		done <- f()
	}(ctx)

	select {
	case res := <-done:
		return res, nil
	case <-ctx.Done():
		return nil, errtrace.Errorf("Func timed out after %d secs", time.Now().Unix()-start.Unix())
	}
}

type ProxyAuth struct {
	Username string
	Password string
}

func (a *ProxyAuth) String() string {
	return a.Username + ":" + a.Password
}

func (a *ProxyAuth) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(a.String()))
}

func (a *ProxyAuth) VerifyBasic(basic string) bool {
	basic = strings.ToLower(basic)
	basic = strings.TrimPrefix(basic, "basic")
	basic = strings.TrimLeft(basic, " ")

	return basic == a.Base64()
}
