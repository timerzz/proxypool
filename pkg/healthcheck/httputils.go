package healthcheck

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	C "github.com/metacubex/mihomo/constant"
	"github.com/timerzz/proxypool/pkg/proxy"
)

// DO NOT EDIT. Copied from clash because it's an unexported function
func urlToMetadata(rawURL string) (addr C.Metadata, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			err = fmt.Errorf("%s scheme not Support", rawURL)
			return
		}
	}
	pi, _ := strconv.Atoi(port)
	addr = C.Metadata{
		Host:    u.Hostname(),
		DstPort: uint16(pi),
	}
	return
}

func HTTPGetViaProxy(clashProxy C.Proxy, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DelayTimeout)
	defer cancel()

	addr, err := urlToMetadata(url)
	if err != nil {
		return err
	}
	conn, err := clashProxy.DialContext(ctx, &addr) // 建立到proxy server的connection，对Proxy的类别做了自适应相当于泛型
	if err != nil {
		return err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		// Note: Dial specifies the dial function for creating unencrypted TCP connections.
		// When httpClient sets this transport, it will use the tcp/udp connection returned from
		// function Dial instead of default tcp/udp connection. It's the key to set custom proxy for http transport
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func HTTPHeadViaProxy(clashProxy C.Proxy, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), DelayTimeout)
	defer cancel()

	addr, err := urlToMetadata(url)
	if err != nil {
		return err
	}
	conn, err := clashProxy.DialContext(ctx, &addr) // 建立到proxy server的connection，对Proxy的类别做了自适应相当于泛型
	if err != nil {
		return err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		// Note: Dial specifies the dial function for creating unencrypted TCP connections.
		// When httpClient sets this transport, it will use the tcp/udp connection returned from
		// function Dial instead of default tcp/udp connection. It's the key to set custom proxy for http transport
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	req.Header = http.Header{
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"},
		"sec-fetch-site":            []string{"same-origin"},
		"upgrade-insecure-requests": []string{"1"},
		"sec-fetch-dest":            []string{"empty"},
		"sec-fetch-mode":            []string{"navigate"},
		"Accept":                    []string{"*/*"},
		"Accept-Encoding":           []string{"gzip, deflate, br"},
		"Connection":                []string{"keep-alive"},
	}
	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%d %s for proxy %s %s", resp.StatusCode, resp.Status, clashProxy.Name(), clashProxy.Addr())
	}
	resp.Body.Close()
	return nil
}

func HTTPGetBodyViaProxy(clashProxy C.Proxy, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DelayTimeout)
	defer cancel()

	addr, err := urlToMetadata(url)
	if err != nil {
		return nil, err
	}
	conn, err := clashProxy.DialContext(ctx, &addr) // 建立到proxy server的connection，对Proxy的类别做了自适应相当于泛型
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		// Note: Dial specifies the dial function for creating unencrypted TCP connections.
		// When httpClient sets this transport, it will use the tcp/udp connection returned from
		// function Dial instead of default tcp/udp connection. It's the key to set custom proxy for http transport
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read speedtest config file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func HTTPGetBodyViaProxyWithTime(clashProxy C.Proxy, url string, t time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	addr, err := urlToMetadata(url)
	if err != nil {
		return nil, err
	}
	conn, err := clashProxy.DialContext(ctx, &addr) // 建立到proxy server的connection，对Proxy的类别做了自适应相当于泛型
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		// Note: Dial specifies the dial function for creating unencrypted TCP connections.
		// When httpClient sets this transport, it will use the tcp/udp connection returned from
		// function Dial instead of default tcp/udp connection. It's the key to set custom proxy for http transport
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read speedtest config file
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Get body without return to save memory
func HTTPGetBodyViaProxyWithTimeNoReturn(clashProxy C.Proxy, url string, t time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	addr, err := urlToMetadata(url)
	if err != nil {
		return err
	}
	conn, err := clashProxy.DialContext(ctx, &addr) // 建立到proxy server的connection，对Proxy的类别做了自适应相当于泛型
	if err != nil {
		return err
	}
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		// Note: Dial specifies the dial function for creating unencrypted TCP connections.
		// When httpClient sets this transport, it will use the tcp/udp connection returned from
		// function Dial instead of default tcp/udp connection. It's the key to set custom proxy for http transport
		Dial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read speedtest config file
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func checkErrorProxies(proxies []proxy.Proxy) bool {
	if proxies == nil {
		return false
	}
	if len(proxies) == 0 {
		return false
	}
	if proxies[0] == nil {
		return false
	}
	return true
}
