package grpc_server

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func BulkInsert(db *gorm.DB, tableName, cols, path string) *gorm.DB {
	query := `
		LOAD DATA LOCAL INFILE %q
		INTO TABLE %s
		FIELDS TERMINATED BY '\t'
		LINES TERMINATED BY '\n' 
		(%s)`
	query = fmt.Sprintf(query, path, tableName, cols)
	return db.Exec(query)
}
