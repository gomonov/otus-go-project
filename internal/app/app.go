package app

import (
	"time"

	"github.com/gomonov/otus-go-project/internal/domain"
	"github.com/gomonov/otus-go-project/internal/storage"
)

type App struct {
	logger  Logger
	storage storage.Storage
	cache   *IPListsCache
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
}

func New(logger Logger, storage storage.Storage, cacheTTL time.Duration) *App {
	return &App{logger: logger, storage: storage, cache: newIPListsCache(cacheTTL)}
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

func (a *App) CheckIPAccess(ip string) (domain.AuthResponse, error) {
	if a.cache.needsReload() {
		a.logger.Debug("Reloading IP lists cache")

		// Загружаем данные из базы
		blacklist, err := a.storage.Subnet().GetByListType(domain.Blacklist)
		if err != nil {
			return domain.AuthResponse{}, err
		}

		whitelist, err := a.storage.Subnet().GetByListType(domain.Whitelist)
		if err != nil {
			return domain.AuthResponse{}, err
		}

		// Обновляем кеш
		if err := a.cache.reload(blacklist, whitelist); err != nil {
			return domain.AuthResponse{}, err
		}

		a.logger.Info("IP lists cache reloaded",
			"blacklist_count", len(blacklist),
			"whitelist_count", len(whitelist))
	}

	status, err := a.cache.checkIP(ip)
	if err != nil {
		return domain.AuthResponse{}, err
	}

	response := domain.AuthResponse{OK: status}
	a.logger.Debug("IP check completed", "ip", ip, "result", status)

	return response, nil
}
