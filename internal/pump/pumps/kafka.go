package pumps

import (
	"context"
	"crypto"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"

	"github.com/onexdemo/iamv2/internal/pump/analytics"
)

// KafkaConf configures the Kafka pump.
type KafkaConf struct {
	Broker                []string          `mapstructure:"broker"`
	ClientID              string            `mapstructure:"client_id"`
	Topic                 string            `mapstructure:"topic"`
	SSLCertFile           string            `mapstructure:"ssl_cert_file"`
	SSLKeyFile            string            `mapstructure:"ssl_key_file"`
	SASLMechanism         string            `mapstructure:"sasl_mechanism"`
	Username              string            `mapstructure:"sasl_username"`
	Password              string            `mapstructure:"sasl_password"`
	Algorithm             string            `mapstructure:"sasl_algorithm"`
	Timeout               time.Duration     `mapstructure:"timeout"`
	MetaData              map[string]string `mapstructure:"meta_data"`
	Compressed            bool              `mapstructure:"compressed"`
	UseSSL                bool              `mapstructure:"use_ssl"`
	SSLInsecureSkipVerify bool              `mapstructure:"ssl_insecure_skip_verify"`
}

// KafkaPump writes records to Kafka.
type KafkaPump struct {
	conf   *KafkaConf
	writer *kafka.Writer
	CommonPumpConfig
}

// New returns a fresh Kafka pump.
func (k *KafkaPump) New() Pump {
	return &KafkaPump{}
}

// GetName returns the pump name.
func (*KafkaPump) GetName() string {
	return "Kafka Pump"
}

// Init initializes the Kafka writer.
func (k *KafkaPump) Init(config any) error {
	k.conf = &KafkaConf{}
	if err := mapstructure.Decode(config, k.conf); err != nil {
		return fmt.Errorf("decode kafka configuration: %w", err)
	}

	if len(k.conf.Broker) == 0 {
		return fmt.Errorf("kafka broker is required")
	}

	if k.conf.Topic == "" {
		return fmt.Errorf("kafka topic is required")
	}

	if k.conf.Timeout == 0 {
		k.conf.Timeout = 10 * time.Second
	}

	tlsConfig, err := k.tlsConfig()
	if err != nil {
		return err
	}

	mechanism, err := k.saslMechanism()
	if err != nil {
		return err
	}

	k.writer = &kafka.Writer{
		Addr:     kafka.TCP(k.conf.Broker...),
		Topic:    k.conf.Topic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			ClientID:    k.conf.ClientID,
			DialTimeout: k.conf.Timeout,
			TLS:         tlsConfig,
			SASL:        mechanism,
		},
		WriteTimeout: k.conf.Timeout,
		ReadTimeout:  k.conf.Timeout,
	}
	if k.conf.Compressed {
		k.writer.Compression = kafka.Snappy
	}

	slog.Info("initialized kafka pump", "topic", k.conf.Topic, "brokers", k.conf.Broker)

	return nil
}

func (k *KafkaPump) tlsConfig() (*tls.Config, error) {
	if !k.conf.UseSSL {
		return nil, nil
	}

	if k.conf.SSLCertFile != "" && k.conf.SSLKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(k.conf.SSLCertFile, k.conf.SSLKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load kafka client certificate: %w", err)
		}

		return &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: k.conf.SSLInsecureSkipVerify,
		}, nil
	}

	return &tls.Config{InsecureSkipVerify: k.conf.SSLInsecureSkipVerify}, nil
}

func (k *KafkaPump) saslMechanism() (sasl.Mechanism, error) {
	switch k.conf.SASLMechanism {
	case "":
		return nil, nil
	case "PLAIN", "plain":
		return plain.Mechanism{Username: k.conf.Username, Password: k.conf.Password}, nil
	case "SCRAM", "scram":
		algorithm := scram.SHA256
		if k.conf.Algorithm == "sha-512" || k.conf.Algorithm == crypto.SHA512.String() {
			algorithm = scram.SHA512
		}

		mechanism, err := scram.Mechanism(algorithm, k.conf.Username, k.conf.Password)
		if err != nil {
			return nil, fmt.Errorf("initialize kafka scram mechanism: %w", err)
		}

		return mechanism, nil

	default:
		return nil, fmt.Errorf("unsupported kafka sasl mechanism %q", k.conf.SASLMechanism)
	}
}

// WriteData publishes a batch of records.
func (k *KafkaPump) WriteData(ctx context.Context, data []any) error {
	if k.writer == nil {
		return fmt.Errorf("kafka pump is not initialized")
	}

	start := time.Now()
	messages := make([]kafka.Message, 0, len(data))

	for _, item := range data {
		record, ok := item.(analytics.Record)
		if !ok {
			continue
		}

		message := map[string]any{
			"timestamp":  record.TimeStamp,
			"username":   record.Username,
			"effect":     record.Effect,
			"conclusion": record.Conclusion,
			"request":    record.Request,
			"policies":   record.Policies,
			"deciders":   record.Deciders,
			"expireAt":   record.ExpireAt,
		}
		for key, value := range k.conf.MetaData {
			message[key] = value
		}

		payload, err := json.Marshal(message)
		if err != nil {
			slog.WarnContext(ctx, "failed to marshal kafka message", "error", err)

			continue
		}

		messages = append(messages, kafka.Message{Time: time.Now(), Value: payload})
	}

	if err := k.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("write kafka messages: %w", err)
	}

	slog.DebugContext(ctx, "wrote kafka records", "count", len(messages), "elapsed", time.Since(start))

	return nil
}
