package sql

// Migration represents a migration in the database.
type Migration struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"uniqueIndex"`
	Batch int
}
