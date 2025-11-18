package app

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gomonov/otus-go-project/internal/domain"
	"github.com/yl2chen/cidranger"
)

type IPListsCache struct {
	mu            sync.RWMutex
	blacklist     cidranger.Ranger
	whitelist     cidranger.Ranger
	lastLoaded    time.Time
	ttl           time.Duration
	isInitialized bool
}

func newIPListsCache(ttl time.Duration) *IPListsCache {
	return &IPListsCache{
		blacklist:     cidranger.NewPCTrieRanger(),
		whitelist:     cidranger.NewPCTrieRanger(),
		ttl:           ttl,
		isInitialized: false,
	}
}

func (c *IPListsCache) needsReload() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isInitialized {
		return true
	}

	return time.Since(c.lastLoaded) > c.ttl
}

func (c *IPListsCache) reload(blacklist, whitelist []domain.Subnet) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newBlacklist := cidranger.NewPCTrieRanger()
	newWhitelist := cidranger.NewPCTrieRanger()

	for _, subnet := range blacklist {
		_, network, err := net.ParseCIDR(subnet.CIDR)
		if err != nil {
			return fmt.Errorf("invalid CIDR in blacklist: %s, error: %w", subnet.CIDR, err)
		}
		if err := newBlacklist.Insert(cidranger.NewBasicRangerEntry(*network)); err != nil {
			return fmt.Errorf("failed to insert into blacklist: %w", err)
		}
	}

	for _, subnet := range whitelist {
		_, network, err := net.ParseCIDR(subnet.CIDR)
		if err != nil {
			return fmt.Errorf("invalid CIDR in whitelist: %s, error: %w", subnet.CIDR, err)
		}
		if err := newWhitelist.Insert(cidranger.NewBasicRangerEntry(*network)); err != nil {
			return fmt.Errorf("failed to insert into whitelist: %w", err)
		}
	}

	c.blacklist = newBlacklist
	c.whitelist = newWhitelist
	c.lastLoaded = time.Now()
	c.isInitialized = true

	return nil
}

func (c *IPListsCache) checkIP(ipStr string) (domain.AuthStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isInitialized {
		return domain.AuthUnknown, fmt.Errorf("IP lists not initialized")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return domain.AuthUnknown, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Проверяем IPv4
	if ip.To4() == nil {
		return domain.AuthUnknown, fmt.Errorf("only IPv4 addresses are supported")
	}

	inBlacklist, err := c.blacklist.Contains(ip)
	if err != nil {
		return domain.AuthUnknown, fmt.Errorf("blacklist check failed: %w", err)
	}

	if inBlacklist {
		return domain.AuthDenied, nil
	}

	inWhitelist, err := c.whitelist.Contains(ip)
	if err != nil {
		return domain.AuthUnknown, fmt.Errorf("whitelist check failed: %w", err)
	}

	if inWhitelist {
		return domain.AuthGranted, nil
	}

	return domain.AuthUnknown, nil
}
