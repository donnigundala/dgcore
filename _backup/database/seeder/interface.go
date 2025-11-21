package seeder

// ISeeder defines the interface for the Seeder struct.
type ISeeder interface {
	RunAll() error
	RunOne(name string) error
	RunAllWithTransaction() error
	RunOneWithTransaction(name string) error
	ListSeeders() []string
}
