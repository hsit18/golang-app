package kafka

import (
	"log"
	"sync"
)

var (
	producerInstance *Producer
	once             sync.Once
)

// InitProducer initializes the global Kafka producer instance
func InitProducer() error {
	var err error
	once.Do(func() {
		producerInstance, err = NewProducer()
		if err != nil {
			log.Printf("Failed to initialize Kafka producer: %v", err)
		} else {
			log.Println("Kafka producer initialized successfully")
		}
	})
	return err
}

// GetProducer returns the global Kafka producer instance
func GetProducer() *Producer {
	if producerInstance == nil {
		log.Println("Kafka producer not initialized. Call InitProducer() first.")
		return nil
	}
	return producerInstance
}

// CloseProducer closes the global Kafka producer instance
func CloseProducer() {
	if producerInstance != nil {
		producerInstance.Close()
		log.Println("Kafka producer closed")
	}
}
