package common

import (
	"context"
	"encoding/base64"
	"go-proxy/binary"
	"net/netip"
	"strings"
	"sync"
	"time"

	"braces.dev/errtrace"
	"github.com/oschwald/maxminddb-golang/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	IP_CHECK_HOST = "api.ipify.org"
)

var ip2countryDb *maxminddb.Reader

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

func GetIpCountry(ip netip.Addr) (string, error) {
	if ip2countryDb == nil {
		file, err := binary.BinaryFS.ReadFile("files/ip-to-country.mmdb")
		if err != nil {
			return "", errtrace.Wrap(err)
		}

		ip2countryDb, err = maxminddb.OpenBytes(file)
		if err != nil {
			return "", errtrace.Wrap(err)
		}
	}

	var res struct {
		CountryCode string `maxminddb:"country_code"`
	}
	lookup := ip2countryDb.Lookup(ip)

	if !lookup.Found() {
		return "", errtrace.Errorf("Country code not found")
	}

	err := lookup.Decode(&res)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	code := string(res.CountryCode)
	if code == "" {
		return "", errtrace.Errorf("Empty country code found")
	}

	return code, nil
}

type GlobalDataMutext struct {
	sync.RWMutex
}

func (m *GlobalDataMutext) Unlock() {
	m.RWMutex.Unlock()
	application.Get().Event.Emit("goproxy:data-changed")
}

var DataMutex GlobalDataMutext
