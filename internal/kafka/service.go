package kafka

import (
	"log"
	"os"
	"strings"
	"sync"
)

var (
	producerInstance *Producer
	consumerInstance *Consumer
	once             sync.Once
	consumerOnce     sync.Once
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

// InitConsumer initializes the global Kafka consumer instance
func InitConsumer() error {
	var err error
	consumerOnce.Do(func() {
		// Get topics from environment variable (comma-separated)
		topicsEnv := os.Getenv("KAFKA_CONSUMER_TOPICS")
		if topicsEnv == "" {
			log.Println("KAFKA_CONSUMER_TOPICS not set, consumer will not be started")
			return
		}

		// Parse topics from comma-separated string
		topics := strings.Split(topicsEnv, ",")
		for i, topic := range topics {
			topics[i] = strings.TrimSpace(topic)
		}

		consumerInstance, err = NewConsumer(topics)
		if err != nil {
			log.Printf("Failed to initialize Kafka consumer: %v", err)
			return
		}

		// Start consuming
		err = consumerInstance.Start()
		if err != nil {
			log.Printf("Failed to start Kafka consumer: %v", err)
			return
		}

		log.Printf("Kafka consumer initialized and started for topics: %v", topics)
	})
	return err
}

// GetConsumer returns the global Kafka consumer instance
func GetConsumer() *Consumer {
	if consumerInstance == nil {
		log.Println("Kafka consumer not initialized. Call InitConsumer() first.")
		return nil
	}
	return consumerInstance
}

// CloseConsumer closes the global Kafka consumer instance
func CloseConsumer() {
	if consumerInstance != nil {
		if err := consumerInstance.Stop(); err != nil {
			log.Printf("Error stopping consumer: %v", err)
		}
		log.Println("Kafka consumer closed")
	}
}
