module github.com/modulix-systems/goose-talk

go 1.25.5

require (
	buf.build/gen/go/co3n/goose-proto/grpc/go v1.6.0-20260102203506-171393a19e83.1
	buf.build/gen/go/co3n/goose-proto/protocolbuffers/go v1.36.11-20260118192846-29625ecf5663.1
	github.com/Masterminds/squirrel v1.5.4
	github.com/brianvoe/gofakeit/v7 v7.2.1
	github.com/go-webauthn/webauthn v0.13.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jackc/pgx/v5 v5.8.0
	github.com/modulix-systems/goose-talk/contracts v0.0.0-00010101000000-000000000000
	github.com/modulix-systems/goose-talk/logger v0.0.0-00010101000000-000000000000
	github.com/modulix-systems/goose-talk/postgres v0.0.0-00010101000000-000000000000
	github.com/modulix-systems/goose-talk/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.46.0
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.11
)

require github.com/rs/zerolog v1.34.0 // indirect

replace (
	github.com/modulix-systems/goose-talk/contracts => ../../pkg/contracts
	github.com/modulix-systems/goose-talk/httpclient => ../../pkg/httpclient
	github.com/modulix-systems/goose-talk/logger => ../../pkg/logger
	github.com/modulix-systems/goose-talk/postgres => ../../pkg/postgres
	github.com/modulix-systems/goose-talk/rabbitmq => ../../pkg/rabbitmq
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-webauthn/x v0.1.21 // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modulix-systems/goose-talk/httpclient v0.0.0-00010101000000-000000000000
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/redis/go-redis/v9 v9.17.2
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
