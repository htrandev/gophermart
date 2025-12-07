-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	login text UNIQUE NOT NULL,
	password text NOT NULL,
	created_at timestamp DEFAULT now(),
	current_balance int DEFAULT 0,
	withdrawn int DEFAULT 0
);

CREATE INDEX ON users (login); 

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	number text NOT NULL,
	status order_status DEFAULT 'NEW',
	created_at timestamp DEFAULT now(),
	updated_at timestamp,
	accrual int NOT NULL DEFAULT 0,
	user_id uuid NOT NULL references users(id) ON DELETE CASCADE 
);

CREATE INDEX ON orders (number); 

CREATE TABLE IF NOT EXISTS withdrawals(
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_number text NOT NULL,
	sum int NOT NULL DEFAULT 0,
	created_at timestamp DEFAULT now(),
	user_id uuid NOT NULL references users(id) ON DELETE CASCADE
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS withdrawals;

DROP TYPE IF EXISTS order_status;
-- +goose StatementEnd
