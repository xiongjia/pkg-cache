package proxy

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xiongjia/pkg-cache/pkg/cache"
	"github.com/xiongjia/pkg-cache/pkg/util"
)

var log = slog.Default()

type Proxy struct {
	cfg        *util.Config
	cache      *cache.Cache
	httpClient *http.Client
}

func New(cfg *util.Config) *Proxy {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: 30,
		}).DialContext,
	}

	if cfg.ParentProxy != "" {
		if proxyURL, err := url.Parse(cfg.ParentProxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			if cfg.ParentAuth != "" {
				transport.ProxyConnectHeader = http.Header{}
				transport.ProxyConnectHeader.Set("Proxy-Authorization",
					"Basic "+base64.StdEncoding.EncodeToString([]byte(cfg.ParentAuth)))
			}
		}
	}

	return &Proxy{
		cfg:        cfg,
		cache:      cache.New(cfg.Cache.Directory, cfg.Cache.MaxSizeMB),
		httpClient: &http.Client{Transport: transport},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	case "/cache/stats":
		p.handleStats(w, r)
		return
	}

	p.handleProxy(w, r)
}

func (p *Proxy) handleStats(w http.ResponseWriter, r *http.Request) {
	size, count, err := p.cache.Stats()
	if err != nil {
		log.Error("failed to get cache stats", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Cache Size: %d bytes (%d MB)\nFiles: %d\n", size, size/(1024*1024), count)
}

func (p *Proxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.String()

	// Extract enabled patterns from config
	var patterns []string
	for _, rule := range p.cfg.CacheRules {
		if rule.Enabled {
			patterns = append(patterns, rule.Pattern)
		}
	}

	// Check if we should cache this request
	shouldCache := !p.cfg.BypassCache && p.cache.ShouldCacheByPatterns(targetURL, patterns)

	// Try to serve from cache
	if shouldCache && p.cache.Exists(targetURL) {
		data, err := p.cache.Get(targetURL)
		if err == nil {
			w.Header().Set("X-Cache", "HIT")
			log.Info("cache hit", "url", targetURL)
			w.Write(data)
			return
		}
		log.Warn("failed to read from cache", "url", targetURL, "error", err)
	}

	// Proxy the request
	resp, err := p.httpClient.Do(r)
	if err != nil {
		log.Error("proxy request failed", "url", targetURL, "error", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read response body", "url", targetURL, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	// Cache if needed
	if shouldCache && resp.StatusCode == http.StatusOK {
		if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			if err := p.cache.Put(targetURL, body); err != nil {
				log.Error("failed to cache response", "url", targetURL, "error", err)
			} else {
				log.Info("cached response", "url", targetURL, "size", len(body))
			}
		}
	}
}
