module github.com/sunnyyssh/designing-software-cw3/order

go 1.24.3

replace github.com/sunnyyssh/designing-software-cw3/shared => ../shared

require (
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v5 v5.7.5
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/sunnyyssh/designing-software-cw3/shared v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)
