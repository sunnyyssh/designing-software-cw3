module github.com/sunnyyssh/designing-software-cw3/payment

go 1.24.3

replace github.com/sunnyyssh/designing-software-cw3/shared => ../shared

require github.com/sunnyyssh/designing-software-cw3/shared v0.0.0-00010101000000-000000000000

require (
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v5 v5.7.5
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
