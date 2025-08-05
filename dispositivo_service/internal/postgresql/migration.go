package postgresql

import (
	"context"
	"fmt"
	"os"
)

func (db DB) Migration() error {
	// Ejecutar migration.sql
	sqlMigrationBytes, err := os.ReadFile("./sql/migrations.sql")
	if err != nil {
		return fmt.Errorf("no se pudo leer el archivo de migración: %w", err)
	}

	sqlMigrationContent := string(sqlMigrationBytes)
	_, err = db.Pool.Exec(context.Background(), sqlMigrationContent)
	if err != nil {
		return fmt.Errorf("error al ejecutar la migración SQL: %w", err)
	}

	logger.Info("✅ Migración ejecutada correctamente.")
	return nil
}
