// repository/tracelog_repository.go
package repository

import (
	"api-gateway/model"
	"database/sql"
)

type TracelogRepository interface {
	Insert(*model.Tracelog) error
}

type tracelogRepository struct {
	db *sql.DB
}

func NewTracelogRepository(db *sql.DB) TracelogRepository {
	return &tracelogRepository{db: db}
}

func (r *tracelogRepository) Insert(m *model.Tracelog) error {
	stmt, err := r.db.Prepare(`
		INSERT IGNORE INTO tracelogs (
			ip, proses, ca_code, product_type, log, tracetime
		) VALUES (
			USER(), ?, ?, ?, ?, CONVERT_TZ(NOW(), '+00:00', '+07:00')
		)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.Proses, m.CaCode, m.ProductType, m.Log)
	return err
}
