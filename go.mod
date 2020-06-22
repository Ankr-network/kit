module github.com/Ankr-network/kit

go 1.14

replace kit.self v0.0.0 => ./

require (
	github.com/caarlos0/env/v6 v6.2.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis v6.15.8+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/jmoiron/sqlx v1.2.0
	github.com/rs/cors v1.7.0
	github.com/shopspring/decimal v1.2.0
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.5.1
	go.mongodb.org/mongo-driver v1.3.4
	go.uber.org/atomic v1.6.0
	go.uber.org/zap v1.15.0
	google.golang.org/genproto v0.0.0-20200620020550-bd6e04640131
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	kit.self v0.0.0
)
