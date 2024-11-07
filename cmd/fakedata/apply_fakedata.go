package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Silverman143/character-service/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	var fakedataPath  string
	flag.StringVar(&fakedataPath, "fakedata-path", "", "path to fake data")

	cfg := config.MustLoad()

	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s",
        cfg.PgSql.User, cfg.PgSql.DbName, cfg.PgSql.SSLMode, cfg.PgSql.Password, cfg.PgSql.Host)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Errorf("err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Чтение и выполнение SQL-файлов
	err = applyFakeData(db, fakedataPath)
	if err != nil {
		log.Fatalf("Failed to apply fake data: %v", err)
	}

	fmt.Println("Fake data applied successfully")
}

func applyFakeData(db *sqlx.DB, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			filePath := filepath.Join(dir, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", file.Name(), err)
			}

			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("failed to execute SQL from %s: %w", file.Name(), err)
			}

			fmt.Printf("Applied %s\n", file.Name())
		}
	}

	return nil
}