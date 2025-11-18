package sqlstorage

import (
	"github.com/gomonov/otus-go-project/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SubnetRepository struct {
	db *sqlx.DB
}

type subnetDB struct {
	ListType string `db:"list_type"`
	CIDR     string `db:"cidr"`
}

func (s subnetDB) toDomain() domain.Subnet {
	return domain.Subnet{
		ListType: domain.ListType(s.ListType),
		CIDR:     s.CIDR,
	}
}

func toSubnetDB(s domain.Subnet) subnetDB {
	return subnetDB{
		ListType: string(s.ListType),
		CIDR:     s.CIDR,
	}
}

func (r *SubnetRepository) Create(subnet *domain.Subnet) error {
	query := `
		INSERT INTO subnets (list_type, cidr) 
		VALUES (:list_type, :cidr)
		ON CONFLICT (list_type, cidr) DO NOTHING
	`

	subnetDB := toSubnetDB(*subnet)
	_, err := r.db.NamedExec(query, &subnetDB)
	return err
}

func (r *SubnetRepository) Delete(listType domain.ListType, cidr string) error {
	query := `DELETE FROM subnets WHERE list_type = $1 AND cidr = $2`

	result, err := r.db.Exec(query, string(listType), cidr)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrSubnetNotFound
	}

	return nil
}

func (r *SubnetRepository) GetByListType(listType domain.ListType) ([]domain.Subnet, error) {
	query := `SELECT list_type, cidr FROM subnets WHERE list_type = $1 ORDER BY cidr`

	var subnetsDB []subnetDB
	err := r.db.Select(&subnetsDB, query, string(listType))
	if err != nil {
		return nil, err
	}

	subnets := make([]domain.Subnet, len(subnetsDB))
	for i, subnet := range subnetsDB {
		subnets[i] = subnet.toDomain()
	}

	return subnets, nil
}
