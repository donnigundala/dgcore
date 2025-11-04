package example

import (
	"github.com/donnigundala/dgcore/migration"
	"gorm.io/gorm"
)

func init() {
	migration.Register(&CreateUsersTable{})
}

// CreateUsersTable creates the users table.
type CreateUsersTable struct{}

// Name returns the name of the migration.
func (m *CreateUsersTable) Name() string {
	return "20231027_create_users_table"
}

// Up creates the users table.
func (m *CreateUsersTable) Up(tx *gorm.DB) error {
	return tx.Exec("CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(255))").Error
}

// Down drops the users table.
func (m *CreateUsersTable) Down(tx *gorm.DB) error {
	return tx.Exec("DROP TABLE users").Error
}

var _ migration.SQLMigration = (*CreateUsersTable)(nil)
