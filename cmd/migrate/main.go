// Command migrate applies SQL migrations from the migrations/ directory.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	_ "github.com/ChiragChiranjib/mcp-proxy/migrations"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	flags   = flag.NewFlagSet("goose", flag.ExitOnError)
	dirFlag = flags.String("dir", "migrations", "Directory with migration files")
	verbose = flags.Bool("v", false, "Enable verbose mode")
)

func main() {
	cfg, err := cfgpkg.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Build DSN from config and open via GORM, then get *sql.DB for goose
	host := cfg.DB.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.DB.Port
	if port == 0 {
		port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", cfg.DB.Username, cfg.DB.Password, host, port, cfg.DB.Name)
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm open: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatalf("gorm sql db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	run(sqlDB, *dirFlag)
}

func run(db *sql.DB, dir string) {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flags: %v", err)
	}
	args := flags.Args()
	if *verbose {
		goose.SetVerbose(true)
	}

	if len(args) < 1 {
		if cmd := os.Getenv("MIGRATION_CMD"); cmd != "" {
			args = append(args, cmd)
		} else {
			flags.Usage()
			return
		}
	}

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("dialect: %v", err)
	}
	command := args[0]
	arguments := []string{}
	if len(args) > 1 {
		arguments = append(arguments, args[1:]...)
	}

	switch command {
	case "create":
		if err := goose.Run("create", nil, dir, arguments...); err != nil {
			log.Fatalf("goose: %v", err)
		}
		return
	case "fix":
		if err := goose.Run("fix", nil, dir); err != nil {
			log.Fatalf("goose: %v", err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := goose.UpContext(ctx, db, dir); err != nil {
		log.Fatalf("goose up: %v", err)
	}
}

func usage() {
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var usageCommands = `
Commands:
  up                   Migrate the DB to the most recent version available
  up-to VERSION        Migrate the DB to a specific VERSION
  down                 Roll back the version by 1
  down-to VERSION      Roll back to a specific VERSION
  redo                 Re-run the latest migration
  reset                Roll back all migrations
  status               Dump the migration status for the current DB
  version              Print the current version of the database
  create NAME          Creates new migration file with the current timestamp
  fix                  Apply sequential ordering to migrations
`
