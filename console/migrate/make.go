package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

const migrationTemplate = `package example // TODO: Ganti 'example' dengan package migrasi Anda

import (
	"context"
	"gorm.io/gorm"
	// "go.mongodb.org/mongo-driver/mongo" // Uncomment jika ini migrasi MongoDB
)

// {{.FuncName}} adalah fungsi untuk migrasi {{.MigrationName}}.
func {{.FuncName}}(ctx context.Context, db any) error {
	// Lakukan type assertion ke koneksi DB yang benar
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		// atau *mongo.Database untuk MongoDB
		panic("{{.FuncName}} expects a *gorm.DB connection")
	}
	
	// TODO: Implementasikan logika migrasi Anda di sini.
	// Contoh: return gormDB.WithContext(ctx).AutoMigrate(&YourModel{})
	
	return nil
}
`

type templateData struct {
	FuncName      string
	MigrationName string
	Timestamp     string
}

type MakeMigrationCommand struct{}

func (c *MakeMigrationCommand) Signature() string {
	return "make:migration [name]"
}

func (c *MakeMigrationCommand) Description() string {
	return "Create a new migration file"
}

func (c *MakeMigrationCommand) Configure(cmd *cobra.Command) {
	cmd.Args = cobra.ExactArgs(1)
}

func (c *MakeMigrationCommand) Handle(cmd *cobra.Command, args []string) error {
	migrationName := args[0]
	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("%s_%s.go", timestamp, migrationName)

	// Konversi snake_case ke CamelCase untuk nama fungsi
	funcNameParts := strings.Split(migrationName, "_")
	for i, part := range funcNameParts {
		funcNameParts[i] = strings.Title(part)
	}
	funcName := strings.Join(funcNameParts, "")

	data := templateData{
		FuncName:      funcName,
		MigrationName: migrationName,
		Timestamp:     timestamp,
	}

	// Determine the migration directory relative to the project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	// This path needs to be adjusted based on where the CLI is run from
	// For now, let's assume the CLI is run from the project root
	migrationDir := filepath.Join(currentDir, "database", "migrations")

	// Ensure the directory exists
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	filePath := filepath.Join(migrationDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	tmpl, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse migration template: %w", err)
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Migration file created: %s\n", filePath)
	return nil
}
