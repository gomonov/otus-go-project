package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gomonov/otus-go-project/internal/domain"
	"github.com/gomonov/otus-go-project/internal/ratelimit"
	"github.com/gomonov/otus-go-project/internal/storage"
)

type App struct {
	logger      Logger
	storage     storage.Storage
	cache       *IPListsCache
	rateLimiter *ratelimit.RateLimiter
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

func New(logger Logger, storage storage.Storage, cacheTTL time.Duration, rateLimiter *ratelimit.RateLimiter) *App {
	return &App{logger: logger, storage: storage, cache: newIPListsCache(cacheTTL), rateLimiter: rateLimiter}
}

func (a *App) CreateSubnet(subnet *domain.Subnet) error {
	a.logger.Info("Creating subnet: ", subnet.CIDR, " for list: ", subnet.ListType)
	return a.storage.Subnet().Create(subnet)
}

func (a *App) DeleteSubnet(listType domain.ListType, cidr string) error {
	a.logger.Info("Deleting subnet: ", cidr, " from list: ", listType)
	return a.storage.Subnet().Delete(listType, cidr)
}

func (a *App) GetSubnetsByListType(listType domain.ListType) ([]domain.Subnet, error) {
	a.logger.Debug("Getting subnets for list: ", listType)
	return a.storage.Subnet().GetByListType(listType)
}

func (a *App) CheckAuth(req domain.AuthRequest) (domain.AuthResponse, error) {
	ipStatus, err := a.checkIPInLists(req.IP)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	if ipStatus == domain.IPInBlacklist {
		a.logger.Info("IP blocked by blacklist", "ip", req.IP)
		return domain.AuthResponse{OK: false}, nil
	}

	if ipStatus == domain.IPInWhitelist {
		a.logger.Info("IP allowed by whitelist", "ip", req.IP)
		return domain.AuthResponse{OK: true}, nil
	}

	if err := a.rateLimiter.Check(context.Background(), req.Login, req.Password, req.IP); err != nil {
		a.logger.Warn("Rate limit exceeded",
			"login", req.Login,
			"ip", req.IP,
			"error", err.Error())
		return domain.AuthResponse{OK: false}, nil
	}

	a.logger.Info("Auth request allowed",
		"login", req.Login,
		"ip", req.IP)
	return domain.AuthResponse{OK: true}, nil
}

func (a *App) checkIPInLists(ip string) (domain.IPListStatus, error) {
	if a.cache.needsReload() {
		a.logger.Debug("Reloading IP lists cache")

		blacklist, err := a.storage.Subnet().GetByListType(domain.Blacklist)
		if err != nil {
			return domain.IPNotInList, err
		}

		whitelist, err := a.storage.Subnet().GetByListType(domain.Whitelist)
		if err != nil {
			return domain.IPNotInList, err
		}

		if err := a.cache.reload(blacklist, whitelist); err != nil {
			return domain.IPNotInList, err
		}

		a.logger.Info("IP lists cache reloaded",
			"blacklist_count", len(blacklist),
			"whitelist_count", len(whitelist))
	}

	return a.cache.checkIP(ip)
}

func (a *App) ResetBuckets(req domain.ResetBucketsRequest) (domain.ResetBucketsResponse, error) {
	if req.Login == "" && req.IP == "" {
		return domain.ResetBucketsResponse{Reset: false},
			fmt.Errorf("either login or ip must be provided")
	}

	err := a.rateLimiter.ResetBuckets(context.Background(), req.Login, req.IP)
	if err != nil {
		a.logger.Error("Failed to reset buckets",
			"login", req.Login,
			"ip", req.IP,
			"error", err.Error())
		return domain.ResetBucketsResponse{Reset: false}, err
	}

	a.logger.Info("Buckets reset successfully",
		"login", req.Login,
		"ip", req.IP)

	return domain.ResetBucketsResponse{Reset: true}, nil
}
