package services

import (
	"api-gateway/model"
	"api-gateway/repository"
	"fmt"
	"log"
	"os"
	"time"
)

type TracelogServices interface {
	Log(proses, ca, product, message string)
}

type tracelogServices struct {
	repo repository.TracelogRepository
}

func NewTracelogServices(s repository.TracelogRepository) TracelogServices {
	return &tracelogServices{repo: s}
}

func (s *tracelogServices) Log(proses, ca, product, message string) {
	logEntry := &model.Tracelog{
		Proses:      proses,
		CaCode:      ca,
		ProductType: product,
		Log:         message,
	}

	err := s.repo.Insert(logEntry)

	// Fallback logging remains the same.
	if err != nil {
		logToFile(fmt.Sprintf("DB insert failed: %v", err))
		logToFile(fmt.Sprintf("[%s] [%s] [%s] %s", time.Now().Format(time.RFC3339), proses, product, message))
	}
}

func logToFile(entry string) {
	_ = os.MkdirAll("logs", os.ModePerm)
	file, err := os.OpenFile("logs/tracelog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open tracelog file: %v", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", 0)
	logger.Println(entry)
}
