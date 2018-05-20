package main

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/juju/errors"
	"github.com/ngaut/log"
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
	KafkaClient sarama.SyncProducer
}

//CreateKafkaProducer creates a new SyncProducer using the given broker addresses and configuration
func (r *Run) CreateKafkaProducer(addrs []string) error {
	var err error

	for i := 0; i < maxRetry; i++ {
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		config.Producer.RequiredAcks = sarama.WaitForAll

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
func (r *Run) TransferData(alertname, env, instance, level, summary string) {
	kafkaMsg := &KafkaMsg{
		Title:       alertname,
		Source:      env,
		Instance:    instance,
		Description: summary,
		Time:        time.Now().Format(timeFormat),
		Level:       level,
		Summary:     summary,
		Expr:        "",
		Value:       "",
		URL:         "",
	}

	alertByte, err := json.Marshal(kafkaMsg)
	if err != nil {
		log.Errorf("Failed to marshal KafkaMsg: %v", err)
		return
	}

	for i := 0; i < maxRetry; i++ {
		if err := r.PushKafkaMsg(string(alertByte)); err != nil {
			log.Errorf("Failed to produce message to kafka cluster: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		return
	}
}

//Scheduler probes services tcp port
func (r *Run) Scheduler() {
	log.Infof("tcp_prober config: %+v", probeConfig)
	for {
		for _, attr := range probeConfig.Service {
			aliveStatus := probeTCP(attr.Addr)
			if !aliveStatus {
				log.Errorf("Failed to dial %s, alert summary is %s", attr.Addr, attr.Summary)
				r.TransferData(attr.Alertname, *clusterName, attr.Addr, attr.Level, attr.Summary)
			}
		}

		time.Sleep(time.Second * 60)
	}
}
