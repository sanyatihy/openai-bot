module github.com/sanyatihy/openai-bot

go 1.19

replace github.com/sanyatihy/openai-go => ../openai-go

require (
	github.com/jackc/pgx/v5 v5.3.1
	github.com/joho/godotenv v1.5.1
	github.com/sanyatihy/openai-go v0.1.0
	go.uber.org/zap v1.24.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
