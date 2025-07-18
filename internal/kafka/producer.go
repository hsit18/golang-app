package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	producer *kafka.Producer
}

// NewProducer creates a new Kafka producer instance
func NewProducer() (*Producer, error) {
	// Kafka configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers":     os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"client.id":             os.Getenv("KAFKA_CLIENT_ID"),
		"acks":                  "all",
		"retries":               3,
		"batch.size":            16384,
		"linger.ms":             1,
		"broker.address.family": "v4",
	}

	// Add security configuration if provided
	if securityProtocol := os.Getenv("KAFKA_SECURITY_PROTOCOL"); securityProtocol != "" {
		config.SetKey("security.protocol", securityProtocol)
	}
	if saslMechanism := os.Getenv("KAFKA_SASL_MECHANISM"); saslMechanism != "" {
		config.SetKey("sasl.mechanism", saslMechanism)
	}
	if saslUsername := os.Getenv("KAFKA_SASL_USERNAME"); saslUsername != "" {
		config.SetKey("sasl.username", saslUsername)
	}
	if saslPassword := os.Getenv("KAFKA_SASL_PASSWORD"); saslPassword != "" {
		config.SetKey("sasl.password", saslPassword)
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

// SendMessage sends a message to the specified topic
func (p *Producer) SendMessage(topic, key, value string) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}

	// Delivery report handler for produced messages
	deliveryChan := make(chan kafka.Event, 1000)

	err := p.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report
	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}

	log.Printf("Message delivered to topic %s [%d] at offset %v\n",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)

	close(deliveryChan)
	return nil
}

// SendMessageAsync sends a message asynchronously to the specified topic
func (p *Producer) SendMessageAsync(topic, key, value string) error {
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}

	err := p.producer.Produce(message, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// SendJSONMessage sends a JSON message to the specified topic
func (p *Producer) SendJSONMessage(topic, key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return p.SendMessage(topic, key, string(jsonData))
}

// SendJSONMessageAsync sends a JSON message asynchronously to the specified topic
func (p *Producer) SendJSONMessageAsync(topic, key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return p.SendMessageAsync(topic, key, string(jsonData))
}

// Flush waits for all messages to be delivered
func (p *Producer) Flush(timeoutMs int) int {
	return p.producer.Flush(timeoutMs)
}

// Close closes the producer
func (p *Producer) Close() {
	p.producer.Close()
}

// GetMetadata returns metadata for all topics
func (p *Producer) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	return p.producer.GetMetadata(topic, allTopics, timeoutMs)
}
