package seeder

import (
	"errors"
	"log"
	"os"

	"gorm.io/gorm"
)

// Package seeder provides a simple and extensible way to manage database
// seeders in Go applications using GORM. It supports:
//
//   - Registering multiple seeders
//   - Defining execution order
//   - Running all or specific seeders with logging
//
// Typical usage:
//
//   db, err := gorm.Open(...)
//   if err != nil {
//       log.Fatal(err)
//   }
//   s := seeder.NewSeeder(db)
//   seeder.SetOrder([]string{"user", "product", "order"})
//   if err := s.RunAll(); err != nil {
//       log.Fatal(err)
//   }
//
// Each seeder function must accept *gorm.DB and return an error.
//
// Example seeder registration (in user_seeder.go):
//
//   func init() {
//       seeder.Register("user", UserSeeder)
//   }
//
// Example seeder implementation:
//
//   func UserSeeder(db *gorm.DB) error {
//       users := []User{
//           {Name: "Admin", Email: "admin@example.com", Password: "password"},
//       }
//       for _, u := range users {
//           if err := db.Where("email = ?", u.Email).FirstOrCreate(&u).Error; err != nil {
//               return err
//           }
//       }
//       return nil
//   }

// Seeder holds the GORM DB instance and a logger for seeding operations.
type Seeder struct {
	db     *gorm.DB
	logger *log.Logger
}

// ISeeder defines the interface for the Seeder struct.
type ISeeder interface {
	RunAll() error
	RunOne(name string) error
	RunAllWithTransaction() error
	RunOneWithTransaction(name string) error
	ListSeeders() []string
	LogPrintf(format string, v ...any)
	LogPrintln(v ...any)
	LogPrint(v ...any)
	LogFatalf(format string, v ...any)
}

// NewSeeder creates a new Seeder instance with the provided DB and a logger.
// The logger outputs to stdout with standard timestamp format.
func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{
		db:     db,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// SeedFunc is the function signature for all seeders.
// Each seeder must accept *gorm.DB and return an error.
type SeedFunc func(*gorm.DB) error

// SetOrder defines the order in which seeders should run.
// If not set, seeders will run in the order they are registered.
//
// Example:
//
//	seeder.SetOrder([]string{"user", "product", "order"})
var seederOrder []string

// seeders map[string]SeedFunc holds the registered seeder functions.
// This map is populated by the Register function.
var seeders = map[string]SeedFunc{}

// Register registers a seeder function with a given name.
// If a seeder with the same name exists, it will be overwritten.
// Recommended to call this in init() of each seeder file.
func Register(name string, fn SeedFunc) {
	seeders[name] = fn
}

// SetOrder defines the order in which seeders should run.
// If not set, seeders will run in the order they are registered.
//
// Example:
//
//	seeder.SetOrder([]string{"user", "product", "order"})
func SetOrder(order []string) {
	seederOrder = order
}

// RunAll runs all registered seeders in the defined order.
// If no order is set, it will run all registered seeders in random map order.
func (s *Seeder) RunAll() error {
	LogPrintln("🚀 Seeder started")

	var names []string
	if len(seederOrder) > 0 {
		names = seederOrder
	} else {
		// fallback to all registered seeder
		for name := range seeders {
			names = append(names, name)
		}
	}

	for _, name := range names {
		if fn, exists := seeders[name]; exists {
			LogPrintf("➡️ Running seeder: %s", name)
			if err := fn(s.db); err != nil {
				LogPrintf("❌ Seeder %s failed: %v", name, err)
				return err
			}
			LogPrintf("✅ Seeder %s completed", name)
		} else {
			LogPrintf("⚠️ Seeder %s not registered", name)
		}
	}

	LogPrintln("🏁 All seeders completed")
	return nil
}

// RunOne executes a specific seeder by name.
// Returns an error if the seeder is not found or fails.
func (s *Seeder) RunOne(name string) error {
	if fn, exists := seeders[name]; exists {
		LogPrintf("➡️ Running seeder: %s", name)
		if err := fn(s.db); err != nil {
			LogPrintf("❌ Seeder %s failed: %v", name, err)
			return err
		}
		LogPrintf("✅ Seeder %s completed", name)
		return nil
	}
	return errors.New("seeder not found")
}

// RunAllWithTransaction executes all seeders inside a single DB transaction.
// If one seeder fails, all changes will be rolled back.
func (s *Seeder) RunAllWithTransaction() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		LogPrintln("🚀 Seeder started (transactional mode)")
		for _, name := range seederOrder {
			LogPrintf("➡️ Running seeder: %s", name)
			if fn, exists := seeders[name]; exists {
				if err := fn(tx); err != nil {
					LogPrintf("❌ Seeder %s failed: %v", name, err)
					return err // rollback
				}
				LogPrintf("✅ Seeder %s completed", name)
			} else {
				LogPrintf("⚠️ Seeder %s not registered", name)
				return errors.New("missing seeder: " + name)
			}
		}
		LogPrintln("🏁 All seeders completed (transaction committed)")
		return nil // commit
	})
}

// RunOneWithTransaction executes a specific seeder by name.
// Returns an error if the seeder is not found or fails.
func (s *Seeder) RunOneWithTransaction(name string) error {
	if fn, exists := seeders[name]; exists {
		return s.db.Transaction(func(tx *gorm.DB) error {
			LogPrintf("➡️ Running seeder: %s (transactional mode)", name)
			if err := fn(tx); err != nil {
				LogPrintf("❌ Seeder %s failed: %v", name, err)
				return err // rollback
			}
			LogPrintf("✅ Seeder %s completed (transaction committed)", name)
			return nil // commit
		})
	}
	return errors.New("seeder not found")
}

// ListSeeders returns a list of registered seeder names in execution order if defined.
func ListSeeders() []string {
	if len(seederOrder) > 0 {
		return seederOrder
	}

	names := make([]string, 0, len(seeders))
	for name := range seeders {
		names = append(names, name)
	}
	return names
}
