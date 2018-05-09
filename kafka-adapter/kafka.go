package main

import (
	//"fmt"
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/ngaut/log"
	"strings"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

//CreateKafkaProduce create a kafka produce
func (r *Run) CreateKafkaProduce() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewManualPartitioner
	var err error
	r.KafkaClient, err = sarama.NewSyncProducer(strings.Split(*kafkaAddress, ","), config)
	return err
}

//PushKafkaMsg push message to kafka cluster
func (r *Run) PushKafkaMsg(msg string) error {
	kafkaMsg := &sarama.ProducerMessage{
		Topic: *kafkaTopic,
		Value: sarama.StringEncoder(msg),
	}
	log.Debugf("get kafka mssage %s", msg)
	_, _, err := r.KafkaClient.SendMessage(kafkaMsg)
	return err
}

//TransferData transfer alert to kafka string
func (r *Run) TransferData(ad *AlertData) {
	for _, at := range ad.Alerts {
		kafkaMsg := &KafkaMsg{
			Title:       getValue(at.Labels, "alertname"),
			Description: getValue(at.Annotations, "description"),
			Expr:        getValue(at.Labels, "expr"),
			Level:       getValue(at.Labels, "level"),
			Node:        getValue(at.Labels, "instance"),
			Source:      getValue(at.Labels, "env"),
			Value:       getValue(at.Annotations, "value"),
			Note:        getValue(at.Annotations, "summary"),
			URL:         at.GeneratorURL,
			Time:        at.StartsAt.Format(timeFormat),
		}
		atByte, err := json.Marshal(kafkaMsg)
		if err != nil {
			log.Errorf("can not marshal data with error %v", err)
			continue
		}

		if err := r.PushKafkaMsg(string(atByte)); err != nil {
			log.Errorf("push message to kafka error %v", err)
		}
	}
}

func getValue(kv KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}
