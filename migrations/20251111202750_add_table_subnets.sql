-- +goose Up
-- +goose StatementBegin
CREATE TYPE list_type AS ENUM ('blacklist', 'whitelist');
CREATE TABLE subnets
(
    list_type list_type NOT NULL,
    cidr CIDR NOT NULL,
    PRIMARY KEY (list_type, cidr)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subnets;
DROP TYPE IF EXISTS list_type;
-- +goose StatementEnd
