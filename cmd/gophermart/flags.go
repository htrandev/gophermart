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

	if addr, ok := os.LookupEnv("ADDRESS"); ok {
		f.addr = addr
	}
	if lvl, ok := os.LookupEnv("LOG_LEVEL"); ok {
		f.logLvl = lvl
	}
	if dbURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		f.databaseURI = dbURI
	}
	if key, ok := os.LookupEnv("KEY"); ok {
		f.key = key
	}
	if ttl, ok := os.LookupEnv("TOKEN_TTL"); ok {
		v, err := strconv.Atoi(ttl)
		if err != nil {
			return f, fmt.Errorf("parse store interval: %w", err)
		}
		f.tokenTTL = time.Duration(v) * time.Second
	}
	if accrualSystemAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		f.accrualSystemAddr = accrualSystemAddr
	}

	return f, nil
}
