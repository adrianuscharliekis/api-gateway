// repository/tracelog_repository.go
package repository

import (
	"api-gateway/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TracelogRepository interface {
	Insert(*model.Tracelog) error
}

type tracelogRepository struct {
	db *gorm.DB
}

func NewTracelogRepository(db *gorm.DB) TracelogRepository {
	return &tracelogRepository{db: db}
}

func (r *tracelogRepository) Insert(m *model.Tracelog) error {
	// This part is already correct.
	ignoreClause := clause.Insert{Modifier: "IGNORE"}

	dataToInsert := map[string]interface{}{
		"ip":           gorm.Expr("user()"),
		"proses":       m.Proses,
		"ca_code":      m.CaCode,
		"product_type": m.ProductType,
		"log":          m.Log,
		"tracetime":    time.Now(),
	}

	result := r.db.Table("tracelogs").
		Clauses(ignoreClause).
		Create(&dataToInsert)

	return result.Error
}
