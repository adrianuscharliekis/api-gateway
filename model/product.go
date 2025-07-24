package model

import "database/sql"

type Product struct {
	ProductId   uint
	ProductName string
	Recid       sql.NullString
	Path        string
}
