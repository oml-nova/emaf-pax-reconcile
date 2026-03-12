package kafka

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	logger "github.com/emaf-pax/pax-reconcile-service/pkg/superlog"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

var (
	kafkaProducerInstance *Producer
	producerOnce         sync.Once
)

// GetKafkaProducer returns a singleton SyncProducer for the configured topic + auth type.
func GetKafkaProducer() (*Producer, error) {
	var producerErr error
	producerOnce.Do(func() {
		topic := os.Getenv("PRODUCER_TOPIC")
		authType := GetAuthType()

		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		config.Producer.Return.Errors = true
		config.Producer.Retry.Max = 3
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.ClientID = "RECONCILE_SERVICE_KAFKA"

		switch strings.ToLower(authType) {
		case "iam":
			producerErr = ConfigureIAM(config)
		default:
			producerErr = ConfigureSASLPlain(config)
		}
		if producerErr != nil {
			return
		}

		producer, err := sarama.NewSyncProducer(GetBrokers(), config)
		if err != nil {
			producerErr = fmt.Errorf("failed to create producer: %w", err)
			return
		}
		kafkaProducerInstance = &Producer{producer: producer, topic: topic}
	})
	if producerErr != nil {
		return nil, producerErr
	}
	return kafkaProducerInstance, nil
}

// ProduceMessage sends a JSON-encoded value to the producer's topic.
func (kp *Producer) ProduceMessage(key string, value interface{}) error {
	if kp == nil {
		return fmt.Errorf("producer is nil")
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}
	msg := &sarama.ProducerMessage{
		Topic:     kp.topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(valueBytes),
		Timestamp: time.Now(),
	}
	_, _, err = kp.producer.SendMessage(msg)
	return err
}

// PublishKafkaMessage is the public generic helper used by application code.
func PublishKafkaMessage(key string, messageType string, messageData interface{}) error {
	producer, err := GetKafkaProducer()
	if err != nil {
		logger.Log().Error("Error getting Kafka producer", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to get producer: %w", err)
	}

	kafkaMsg := map[string]interface{}{
		"type": messageType,
		"data": messageData,
	}

	if err = producer.ProduceMessage(key, kafkaMsg); err != nil {
		logger.Log().Error("Error producing message to Kafka", map[string]interface{}{
			"error": err.Error(), "msg": kafkaMsg,
		})
		return fmt.Errorf("error producing message to Kafka: %w on topic %s", err, producer.topic)
	}
	return nil
}

// Close shuts down the producer.
func (kp *Producer) Close() error {
	if kp != nil && kp.producer != nil {
		return kp.producer.Close()
	}
	return nil
}
