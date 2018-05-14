package main

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
)

const (
	timeFormat    = "2006-01-02 15:04:05"
	maxRetry      = 12
	retryInterval = 5 * time.Second
)

//KafkaMsg represents kafka message
type KafkaMsg struct {
	Title       string `json:"event_object"`
	Source      string `json:"object_name"`
	Instance    string `json:"object_ip"`
	Description string `json:"event_msg"`
	Time        string `json:"event_time"`
	Level       string `json:"event_level"`
	Summary     string `json:"summary"`
	Expr        string `json:"expr"`
	Value       string `json:"value"`
	URL         string `json:"url"`
}

//Run represents runtime information
type Run struct {
	Rdr         *render.Render
	AlertMsgs   chan *AlertData
	KafkaClient sarama.SyncProducer
}

func getValue(kv KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}

//CreateKafkaProducer creates a new SyncProducer using the given broker addresses and configuration
func (r *Run) CreateKafkaProducer(addrs []string) error {
	var err error

	for i := 0; i < maxRetry; i++ {
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		config.Producer.RequiredAcks = sarama.WaitForLocal

		r.KafkaClient, err = sarama.NewSyncProducer(addrs, config)

		if err != nil {
			log.Errorf("create kafka producer with error: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		return nil
	}

	return errors.Trace(err)
}

//PushKafkaMsg pushes message to kafka cluster
func (r *Run) PushKafkaMsg(msg string) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic: *kafkaTopic,
		Value: sarama.StringEncoder(msg),
	}

	partition, offset, err := r.KafkaClient.SendMessage(kafkaMsg)
	if err != nil {
		return errors.Trace(err)
	}
	log.Infof("Produced message %s to kafka cluster partition %d with offset %d", msg, partition, offset)
	return nil
}

//TransferData transfers AlertData to string and sends message to kafka
func (r *Run) TransferData(ad *AlertData) {
	for _, alert := range ad.Alerts {
		kafkaMsg := &KafkaMsg{
			Title:       getValue(alert.Labels, "alertname"),
			Source:      getValue(alert.Labels, "env"),
			Instance:    getValue(alert.Labels, "instance"),
			Description: getValue(alert.Annotations, "description"),
			Time:        alert.StartsAt.Format(timeFormat),
			Level:       getValue(alert.Labels, "level"),
			Summary:     getValue(alert.Annotations, "summary"),
			Expr:        getValue(alert.Labels, "expr"),
			Value:       getValue(alert.Annotations, "value"),
			URL:         alert.GeneratorURL,
		}

		alertByte, err := json.Marshal(kafkaMsg)
		if err != nil {
			log.Errorf("Failed to marshal KafkaMsg: %v", err)
			continue
		}

		if err := r.PushKafkaMsg(string(alertByte)); err != nil {
			log.Errorf("Failed to produce message to kafka cluster: %v", err)
		}
	}
}

//Scheduler for monitoring chan data
func (r *Run) Scheduler() {
	for {
		for alert := range r.AlertMsgs {
			r.TransferData(alert)
		}

		time.Sleep(3 * time.Second)
	}
}
