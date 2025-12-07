package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type flags struct {
	addr              string
	logLvl            string
	databaseURI       string
	key               string
	tokenTTL          time.Duration
	accrualSystemAddr string
}

func parseFlags() (flags, error) {
	var f flags

	flag.StringVar(&f.addr, "a", "localhost:8080", "address to run server")
	flag.StringVar(&f.logLvl, "lvl", "debug", "log level")
	flag.StringVar(&f.databaseURI, "d", "", "db uri")
	flag.StringVar(&f.key, "k", "", "secret key")
	flag.DurationVar(&f.tokenTTL, "ttl", 1*time.Hour, "token ttl in seconds")
	flag.StringVar(&f.accrualSystemAddr, "r", "localhost:8080", "addres of accrual system")

	if addr := os.Getenv("ADDRESS"); addr != "" {
		f.addr = addr
	}
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		f.logLvl = lvl
	}
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		f.databaseURI = dbURI
	}
	if key := os.Getenv("KEY"); key != "" {
		f.key = key
	}
	if ttl := os.Getenv("TOKEN_TTL"); ttl != "" {
		v, err := strconv.Atoi(ttl)
		if err != nil {
			return f, fmt.Errorf("parse store interval: %w", err)
		}
		f.tokenTTL = time.Duration(v) * time.Second
	}
	if accrualSystemAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSystemAddr != "" {
		f.accrualSystemAddr = accrualSystemAddr
	}

	return f, nil
}
