package http_server

import (
	"awesomeProject/config"
	table "awesomeProject/pkg/postgres"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
)

func InitDb(cfg *config.Config) (*sql.DB, *redis.Client, error) {
	portAtoi, _ := strconv.Atoi(cfg.Postgres.Port)
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Postgres.Host, portAtoi, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode)
	db, err := sql.Open(cfg.Postgres.PgDriver, psqlconn)
	if err != nil {
		panic(err)
	}
	err = table.CreateTable(db)
	if err != nil {
		return nil, nil, err
	}
	err = table.CreateTrigger(db)
	if err != nil {
		return nil, nil, err
	}

	i, err := strconv.Atoi(cfg.Redis.DB)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       i,
	})

	return db, redisClient, nil
}
