package dbhelper

import "hrms/database"

func tableExists(tableName string) bool {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName)
	return err == nil && count > 0
}

func columnExists(tableName, columnName string) bool {
	var count int
	err := database.DB.Get(&count, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
	`, tableName, columnName)
	return err == nil && count > 0
}
