package dgseeder

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"gorm.io/gorm"
)

// SeedFunc is the function signature for all seeders.
type SeedFunc func(*gorm.DB) error

// Seeder holds the GORM DB instance and a logger for seeding operations.
type Seeder struct {
	db          *gorm.DB
	logger      *slog.Logger
	seeders     map[string]SeedFunc
	seederOrder []string
}

// New creates a new Seeder instance. It requires a GORM DB instance and accepts
// an optional slog.Logger. If no logger is provided, a default one is created.
func New(db *gorm.DB, logger *slog.Logger) *Seeder {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return &Seeder{
		db:      db,
		logger:  logger.With("component", "seeder"),
		seeders: make(map[string]SeedFunc),
	}
}

// Register registers a seeder function with a given name.
// If a seeder with the same name exists, it will be overwritten.
func (s *Seeder) Register(name string, fn SeedFunc) {
	s.logger.Debug("Registering seeder.", "seeder_name", name)
	s.seeders[name] = fn
}

// SetOrder defines the order in which seeders should run.
// If not set, seeders will run in alphabetical order by name.
func (s *Seeder) SetOrder(order []string) {
	s.seederOrder = order
}

// ListSeeders returns a list of registered seeder names in execution order if defined.
func (s *Seeder) ListSeeders() []string {
	if len(s.seederOrder) > 0 {
		// Return a copy to prevent external modification
		order := make([]string, len(s.seederOrder))
		copy(order, s.seederOrder)
		return order
	}

	// If no order is set, return alphabetically sorted names for deterministic execution.
	names := make([]string, 0, len(s.seeders))
	for name := range s.seeders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RunAll runs all registered seeders in the defined order.
// If no order is set, it will run them in alphabetical order.
func (s *Seeder) RunAll() error {
	s.logger.Info("🚀 Seeder started")
	names := s.ListSeeders()

	for _, name := range names {
		if fn, exists := s.seeders[name]; exists {
			s.logger.Info("➡️ Running seeder", "seeder_name", name)
			if err := fn(s.db); err != nil {
				s.logger.Error("❌ Seeder failed", "seeder_name", name, "error", err)
				return fmt.Errorf("seeder '%s' failed: %w", name, err)
			}
			s.logger.Info("✅ Seeder completed", "seeder_name", name)
		} else {
			s.logger.Warn("⚠️ Seeder specified in order but not registered, skipping.", "seeder_name", name)
		}
	}

	s.logger.Info("🏁 All seeders completed successfully")
	return nil
}

// RunOne executes a specific seeder by name.
// Returns an error if the seeder is not found or fails.
func (s *Seeder) RunOne(name string) error {
	if fn, exists := s.seeders[name]; exists {
		s.logger.Info("➡️ Running seeder", "seeder_name", name)
		if err := fn(s.db); err != nil {
			s.logger.Error("❌ Seeder failed", "seeder_name", name, "error", err)
			return fmt.Errorf("seeder '%s' failed: %w", name, err)
		}
		s.logger.Info("✅ Seeder completed", "seeder_name", name)
		return nil
	}
	s.logger.Error("Seeder not found", "seeder_name", name)
	return fmt.Errorf("seeder '%s' not found", name)
}

// RunAllWithTransaction executes all seeders inside a single DB transaction.
// If one seeder fails, all changes will be rolled back.
func (s *Seeder) RunAllWithTransaction() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		s.logger.Info("🚀 Seeder started (transactional mode)")
		names := s.ListSeeders()

		for _, name := range names {
			s.logger.Info("➡️ Running seeder", "seeder_name", name)
			if fn, exists := s.seeders[name]; exists {
				if err := fn(tx); err != nil {
					s.logger.Error("❌ Seeder failed, rolling back transaction", "seeder_name", name, "error", err)
					return err // rollback
				}
				s.logger.Info("✅ Seeder completed", "seeder_name", name)
			} else {
				s.logger.Warn("⚠️ Seeder specified in order but not registered, skipping.", "seeder_name", name)
			}
		}
		s.logger.Info("🏁 All seeders completed, committing transaction")
		return nil // commit
	})
}

// RunOneWithTransaction executes a specific seeder by name inside a DB transaction.
func (s *Seeder) RunOneWithTransaction(name string) error {
	if fn, exists := s.seeders[name]; exists {
		return s.db.Transaction(func(tx *gorm.DB) error {
			s.logger.Info("➡️ Running seeder (transactional mode)", "seeder_name", name)
			if err := fn(tx); err != nil {
				s.logger.Error("❌ Seeder failed, rolling back transaction", "seeder_name", name, "error", err)
				return err // rollback
			}
			s.logger.Info("✅ Seeder completed, committing transaction", "seeder_name", name)
			return nil // commit
		})
	}
	s.logger.Error("Seeder not found", "seeder_name", name)
	return fmt.Errorf("seeder '%s' not found", name)
}
