// Package sql contains small SQL-related interfaces used by the health checkers.
//
// The package keeps the checker layer independent from *database/sql while still
// making it easy to adapt standard library types.
//
// # Example
//
//	import databasesql "database/sql"
//
//	var db *databasesql.DB
//	check := checker.NewDBChecker(db, 5*time.Second)
//
//	if err := check.Check(context.Background()); err != nil {
//		log.Printf("database unhealthy: %v", err)
//	}
package sql
