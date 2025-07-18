package muxhttphandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hsit18/golang-app/internal/kafka"
)

var myServer *http.Server

func NewServer() {
	// Initialize Kafka producer
	if err := kafka.InitProducer(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka producer: %v", err)
	}

	// Initialize Kafka consumer
	if err := kafka.InitConsumer(); err != nil {
		log.Printf("Warning: Failed to initialize Kafka consumer: %v", err)
	}

	myServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("MUX_HTTP_PORT")),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// Kafka endpoints
	router.HandleFunc("/api/kafka/send", handleKafkaSend).Methods("POST")
	router.HandleFunc("/api/kafka/send-async", handleKafkaSendAsync).Methods("POST")

	// Consumer management endpoints
	router.HandleFunc("/api/kafka/consumer/status", handleConsumerStatus).Methods("GET")

	myServer.Handler = router
	if err := myServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
	log.Println("Stopped serving new connections.")
}

// KafkaMessage represents the structure for Kafka message requests
type KafkaMessage struct {
	Topic   string      `json:"topic"`
	Key     string      `json:"key"`
	Message interface{} `json:"message"`
}

// handleKafkaSend handles synchronous Kafka message sending
func handleKafkaSend(w http.ResponseWriter, r *http.Request) {
	var msg KafkaMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if msg.Topic == "" {
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	producer := kafka.GetProducer()
	if producer == nil {
		http.Error(w, "Kafka producer not available", http.StatusServiceUnavailable)
		return
	}

	var err error
	switch v := msg.Message.(type) {
	case string:
		err = producer.SendMessage(msg.Topic, msg.Key, v)
	default:
		err = producer.SendJSONMessage(msg.Topic, msg.Key, msg.Message)
	}

	if err != nil {
		log.Printf("Failed to send Kafka message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
		"topic":   msg.Topic,
		"key":     msg.Key,
	})
}

// handleKafkaSendAsync handles asynchronous Kafka message sending
func handleKafkaSendAsync(w http.ResponseWriter, r *http.Request) {
	var msg KafkaMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if msg.Topic == "" {
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	producer := kafka.GetProducer()
	if producer == nil {
		http.Error(w, "Kafka producer not available", http.StatusServiceUnavailable)
		return
	}

	var err error
	switch v := msg.Message.(type) {
	case string:
		err = producer.SendMessageAsync(msg.Topic, msg.Key, v)
	default:
		err = producer.SendJSONMessageAsync(msg.Topic, msg.Key, msg.Message)
	}

	if err != nil {
		log.Printf("Failed to send async Kafka message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Message queued for sending",
		"topic":   msg.Topic,
		"key":     msg.Key,
	})
}

// handleConsumerStatus returns the status of the Kafka consumer
func handleConsumerStatus(w http.ResponseWriter, r *http.Request) {
	consumer := kafka.GetConsumer()

	status := map[string]interface{}{
		"consumer_initialized": consumer != nil,
		"timestamp":            time.Now().Format("2006-01-02 15:04:05"),
	}

	if consumer != nil {
		// Get assignment information
		assignment, err := consumer.GetAssignment()
		if err != nil {
			status["assignment_error"] = err.Error()
		} else {
			var assignmentInfo []map[string]interface{}
			for _, tp := range assignment {
				assignmentInfo = append(assignmentInfo, map[string]interface{}{
					"topic":     *tp.Topic,
					"partition": tp.Partition,
					"offset":    tp.Offset,
				})
			}
			status["assignment"] = assignmentInfo
		}

		// Get metadata
		metadata, err := consumer.GetMetadata(5000)
		if err != nil {
			status["metadata_error"] = err.Error()
		} else {
			var topics []string
			for topicName := range metadata.Topics {
				topics = append(topics, topicName)
			}
			status["available_topics"] = topics
			status["broker_count"] = len(metadata.Brokers)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func StopServer(shutdownCtx context.Context) error {
	log.Println("Shutting down the MUX server...")

	// Close Kafka consumer
	kafka.CloseConsumer()

	// Close Kafka producer
	kafka.CloseProducer()

	return myServer.Shutdown(shutdownCtx)
}
