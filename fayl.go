package fayl

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/redhajuanda/fayl/parser"
	"github.com/sirupsen/logrus"
)

type Option struct {
	DB            *sql.DB
	QueryLocation string
	DriverName    string
	Placeholder   parser.Placeholder
	LogLevel      uint32 // 0 = Panic, 1 = Fatal, 2 = Error, 3 = Warn, 4 = Info, 5 = Debug, 6 = Trace
	LogFormatter  string // "json" or "text"
	LogOutput     io.Writer
}

// Init initializes a new fayl client.
func Init(opt Option) (*Client, error) {

	return initFayl(opt)

}

// initFayl initializes a new fayl client with the given options.
func initFayl(opt Option) (*Client, error) {

	// init logger
	log := newLogger("fayl")
	// Set log level
	log.Logger.SetLevel(logrus.Level(opt.LogLevel))
	// Set log formatter
	if opt.LogFormatter == "" {
		opt.LogFormatter = "json" // default to json formatter
	}
	switch opt.LogFormatter {
	case "json":
		log.Logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.Logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	}
	// Set log output
	if opt.LogOutput != nil {
		log.Logger.SetOutput(opt.LogOutput)
	} else {
		log.Logger.SetOutput(os.Stdout)
	}

	db := sqlx.NewDb(opt.DB, opt.DriverName)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Initialize the runners map to store SQL queries
	var runners = make(map[string]string)

	// Check if QueryLocation exists
	if _, err := os.Stat(opt.QueryLocation); os.IsNotExist(err) {
		return nil, fmt.Errorf("query location does not exist: %s", opt.QueryLocation)
	}

	// Walk through the directory and its subdirectories
	err := filepath.Walk(opt.QueryLocation, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Process only .sql files
		if strings.ToLower(filepath.Ext(path)) == ".sql" {
			// Read the SQL file
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}

			// Get the relative path from QueryLocation
			relPath, err := filepath.Rel(opt.QueryLocation, path)
			if err != nil {
				return fmt.Errorf("error getting relative path: %v", err)
			}

			// Create the key by removing the extension and replacing path separators with dots
			key := strings.TrimSuffix(relPath, filepath.Ext(relPath))
			key = strings.ReplaceAll(key, string(os.PathSeparator), ".")

			// Store the SQL query in the runners map
			runners[key] = string(content)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through query location: %v", err)
	}

	return &Client{
		db:          &DB{DB: db},
		runners:     runners,
		placeholder: opt.Placeholder,
		log:         log,
	}, nil

}
