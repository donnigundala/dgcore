package broker

// Config represents the configuration for a single broker connection.
type Config struct {
	Driver   string         `yaml:"driver"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	NATS     NATSConfig     `yaml:"nats"`
}

// KafkaConfig holds the specific configuration for a Kafka connection.
type KafkaConfig struct {
	Brokers  []string `yaml:"brokers"`
	GroupID  string   `yaml:"group_id"`
	// Add other Kafka-specific settings here, e.g., SASL, TLS.
}

// RabbitMQConfig holds the specific configuration for a RabbitMQ connection.
type RabbitMQConfig struct {
	URL      string `yaml:"url"`
	Exchange string `yaml:"exchange"`
	// Add other RabbitMQ-specific settings here.
}

// NATSConfig holds the specific configuration for a NATS connection.
type NATSConfig struct {
	URL string `yaml:"url"`
	// Add other NATS-specific settings here, e.g., credentials.
}
