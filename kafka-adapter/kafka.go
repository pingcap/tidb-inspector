package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

//KafkaMsg represents kafka message
type KafkaMsg struct {
	Title       string `json:"title"`
	Source      string `json:"source"`
	Node        string `json:"node"`
	Expr        string `json:"expr"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Level       string `json:"level"`
	Note        string `json:"note"`
	Value       string `json:"value"`
	Time        string `json:"time"`
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
func (r *Run) CreateKafkaProducer() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewManualPartitioner
	var err error
	r.KafkaClient, err = sarama.NewSyncProducer(strings.Split(*kafkaAddress, ","), config)
	return err
}

//PushKafkaMsg pushes message to kafka cluster
func (r *Run) PushKafkaMsg(msg string) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic: *kafkaTopic,
		Value: sarama.StringEncoder(msg),
	}
	log.Infof("sending message %s to kafka", msg)
	_, _, err := r.KafkaClient.SendMessage(kafkaMsg)
	return err
}

//TransferData transfers AlertData to string and sends message to kafka
func (r *Run) TransferData(ad *AlertData) {
	for _, alert := range ad.Alerts {
		kafkaMsg := &KafkaMsg{
			Title:       getValue(alert.Labels, "alertname"),
			Description: getValue(alert.Annotations, "description"),
			Expr:        getValue(alert.Labels, "expr"),
			Level:       getValue(alert.Labels, "level"),
			Node:        getValue(alert.Labels, "instance"),
			Source:      getValue(alert.Labels, "env"),
			Value:       getValue(alert.Annotations, "value"),
			Note:        getValue(alert.Annotations, "summary"),
			URL:         alert.GeneratorURL,
			Time:        alert.StartsAt.Format(timeFormat),
		}

		alertByte, err := json.Marshal(kafkaMsg)
		if err != nil {
			log.Errorf("can not marshal KafkaMsg with error %v", err)
			continue
		}

		if err := r.PushKafkaMsg(string(alertByte)); err != nil {
			log.Errorf("sending message to kafka with error %v", err)
		}
	}
}

//Scheduler for monitoring chan data
func (r *Run) Scheduler() {
	for {
		lenAlertMsgs := len(r.AlertMsgs)
		if lenAlertMsgs > 0 {
			for i := 0; i < lenAlertMsgs; i++ {
				r.TransferData(<-r.AlertMsgs)
			}
		}
		time.Sleep(3 * time.Second)
	}
}
