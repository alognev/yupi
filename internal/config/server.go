package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultStoreInterval   = 300 * time.Second
	DefaultFileStoragePath = "tmp/metrics-db.json"
	DefaultRestore         = true
)

type ServerConfig struct {
	ServerAddr      string `env:"ADDRESS"`
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	UseGzip         bool `env:"USE_GZIP" envDefault:"true"`
}

// Выставляет значения конфиг из аргументов командной строки
func SetServerConfig() ServerConfig {
	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	a := flag.String("a", DefaultServerAddr, "Адрес сервера")
	flag.Parse()

	if strings.TrimSpace(cfg.ServerAddr) == "" {
		cfg.ServerAddr = *a
	}

	var storeIntervalSeconds int
	flag.IntVar(&storeIntervalSeconds, "i", int(DefaultStoreInterval.Seconds()), "store interval in seconds")
	flag.StringVar(&cfg.FileStoragePath, "f", DefaultFileStoragePath, "file storage path")
	flag.BoolVar(&cfg.Restore, "r", DefaultRestore, "restore metrics from file")

	flag.Parse()

	// Переопределяем, фиксим кейс выше, когда в енвах может быть пустая строка, разрешаем такие значения
	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			storeIntervalSeconds = i
		}
	}

	cfg.StoreInterval = time.Duration(storeIntervalSeconds) * time.Second

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		cfg.FileStoragePath = envPath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		cfg.Restore = envRestore == "true"
	}

	return cfg
}
