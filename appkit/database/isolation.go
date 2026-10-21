// Package database validates the storage boundary for independently deployed applications.
package database

import (
	"fmt"

	"github.com/alexQi/kerthus-core/appidentity"
	driver "github.com/go-sql-driver/mysql"
)

func ExpectedName(code string) (string, error) {
	return appidentity.DatabaseName(code)
}

// ValidateDSN rejects platform, shared, and incorrectly named databases before an app starts.
func ValidateDSN(raw, code string) error {
	expected, err := ExpectedName(code)
	if err != nil {
		return err
	}
	cfg, err := driver.ParseDSN(raw)
	if err != nil || cfg.DBName == "" {
		return fmt.Errorf("invalid application database DSN")
	}
	if cfg.DBName != expected {
		return fmt.Errorf("application %s must use independent database %s", code, expected)
	}
	return nil
}
