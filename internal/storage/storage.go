package storage

import (
	"github.com/gomonov/otus-go-project/internal/domain"
)

type Storage interface {
	Subnet() SubnetRepository
	Close() error
}

type SubnetRepository interface {
	Create(subnet *domain.Subnet) error
	Delete(listType domain.ListType, network string) error
	GetByListType(listType domain.ListType) ([]domain.Subnet, error)
}
