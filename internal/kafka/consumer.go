package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	topics   []string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConsumer creates a new Kafka consumer instance
func NewConsumer(topics []string) (*Consumer, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("at least one topic must be specified")
	}

	// Kafka consumer configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers":       os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"group.id":                os.Getenv("KAFKA_CONSUMER_GROUP_ID"),
		"client.id":               os.Getenv("KAFKA_CLIENT_ID"),
		"auto.offset.reset":       "earliest",
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
		"session.timeout.ms":      30000,
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

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		consumer: consumer,
		topics:   topics,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start begins consuming messages from subscribed topics
func (c *Consumer) Start() error {
	// Subscribe to topics
	err := c.consumer.SubscribeTopics(c.topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics %v: %w", c.topics, err)
	}

	log.Printf("Kafka consumer started, subscribed to topics: %v", c.topics)

	// Start consuming in a goroutine
	go c.consumeLoop()

	return nil
}

// consumeLoop runs the main consumer loop
func (c *Consumer) consumeLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Consumer loop panic recovered: %v", r)
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			log.Println("Consumer loop stopped")
			return
		default:
			// Poll for messages with timeout
			msg, err := c.consumer.ReadMessage(1000 * time.Millisecond)
			if err != nil {
				// Check if it's a timeout (expected during normal operation)
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Consumer error: %v", err)
				continue
			}

			// Process the message
			c.processMessage(msg)
		}
	}
}

// processMessage processes a received message
func (c *Consumer) processMessage(msg *kafka.Message) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Extract message details
	topic := *msg.TopicPartition.Topic
	partition := msg.TopicPartition.Partition
	offset := msg.TopicPartition.Offset
	key := string(msg.Key)
	value := string(msg.Value)

	// Log the received message
	log.Printf("📥 [%s] Received message from topic '%s' [partition: %d, offset: %d]",
		timestamp, topic, partition, offset)

	if key != "" {
		log.Printf("   📋 Key: %s", key)
	}

	log.Printf("   📝 Message: %s", value)

	// Try to pretty-print JSON if possible
	if strings.HasPrefix(strings.TrimSpace(value), "{") || strings.HasPrefix(strings.TrimSpace(value), "[") {
		log.Printf("   📄 Raw JSON: %s", value)
	}

	log.Printf("   ⏰ Headers: %v", c.formatHeaders(msg.Headers))
	log.Println("   " + strings.Repeat("-", 50))
}

// formatHeaders formats message headers for logging
func (c *Consumer) formatHeaders(headers []kafka.Header) string {
	if len(headers) == 0 {
		return "none"
	}

	var headerStrs []string
	for _, header := range headers {
		headerStrs = append(headerStrs, fmt.Sprintf("%s=%s", header.Key, string(header.Value)))
	}

	return strings.Join(headerStrs, ", ")
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	log.Println("Stopping Kafka consumer...")

	// Cancel the context to stop the consume loop
	c.cancel()

	// Close the consumer
	err := c.consumer.Close()
	if err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}

	log.Println("Kafka consumer stopped")
	return nil
}

// GetMetadata returns metadata for subscribed topics
func (c *Consumer) GetMetadata(timeoutMs int) (*kafka.Metadata, error) {
	return c.consumer.GetMetadata(nil, true, timeoutMs)
}

// GetAssignment returns the current topic partition assignment
func (c *Consumer) GetAssignment() ([]kafka.TopicPartition, error) {
	return c.consumer.Assignment()
}

// Commit manually commits the current offset
func (c *Consumer) Commit() error {
	_, err := c.consumer.Commit()
	return err
}
