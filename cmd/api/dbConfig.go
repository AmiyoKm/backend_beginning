package main

import "github.com/AmiyoKm/go-backend/internal/env"

var DBConfig = dbConfig{
	Addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
	maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
	maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
	maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
}

