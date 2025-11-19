package domain

import "errors"

type ListType string

const (
	Blacklist ListType = "blacklist"
	Whitelist ListType = "whitelist"
)

type Subnet struct {
	ListType ListType
	CIDR     string
}

var ErrSubnetNotFound = errors.New("subnet not found")
