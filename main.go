package fayl

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Client struct {
	db      *sqlx.DB
	runners map[string]string
}

type Option struct {
	DB            *sql.DB
	QueryLocation string
	DriverName    string
}

// Init initializes a new fayl client.
func Init(opt Option) (*Client, error) {

	return initFayl(opt)

}

// initFayl initializes a new fayl client with the given options.
func initFayl(opt Option) (*Client, error) {

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
		db:      db,
		runners: runners,
	}, nil
}

func (c *Client) Run(runner string) Runnerer {

	return newRunner(runnerParams{
		runnerCode: runner,
		client:     c,
		// log:           c.log,
		inTransaction: false,
	})
}
