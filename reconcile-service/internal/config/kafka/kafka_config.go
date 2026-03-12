package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
)

// MSKAccessTokenProvider implements sarama.AccessTokenProvider for MSK IAM auth.
type MSKAccessTokenProvider struct {
	Region string
}

func (m *MSKAccessTokenProvider) Token() (*sarama.AccessToken, error) {
	token, _, err := signer.GenerateAuthToken(context.TODO(), m.Region)
	return &sarama.AccessToken{Token: token}, err
}

// ConfigureSASLPlain configures SASL/PLAIN auth on the sarama config.
func ConfigureSASLPlain(config *sarama.Config) error {
	username := os.Getenv("KAFKA_SASL_USERNAME")
	password := os.Getenv("KAFKA_SASL_PASSWORD")
	if username == "" || password == "" {
		return errors.New("KAFKA_SASL_USERNAME and KAFKA_SASL_PASSWORD must be set")
	}
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.User = username
	config.Net.SASL.Password = password
	return nil
}

// ConfigureIAM configures AWS MSK IAM auth (short-lived OAuth tokens).
func ConfigureIAM(config *sarama.Config) error {
	region := os.Getenv("AWS_REGION")
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	config.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: region}
	config.Net.TLS.Enable = true
	return nil
}

// GetBrokers parses the KAFKA_HOST JSON array from environment.
func GetBrokers() []string {
	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		return []string{}
	}
	var brokers []string
	if err := json.Unmarshal([]byte(kafkaHost), &brokers); err != nil {
		panic(fmt.Sprintf("Invalid KAFKA_HOST format: %v", err))
	}
	return brokers
}

// GetAuthType returns the Kafka auth type from environment (defaults to "sasl").
func GetAuthType() string {
	ssl := strings.ToLower(os.Getenv("KAFKA_SSL"))
	if ssl == "true" {
		return "iam"
	}
	return "sasl"
}
