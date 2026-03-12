package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	brokerAddress = "127.0.0.1:9092"
	topic         = "test-topic"
	numPartitions = 3
	replication   = 1
)

// 创建 Topic
func createTopic() {
	controllerConn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		log.Fatal("Failed to connect controller:", err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replication,
	})
	if err != nil {
		log.Println("Topic may already exist:", err)
	} else {
		log.Println("Topic created:", topic)
	}
}

// 生产者函数
func produce() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
	})
	defer writer.Close()

	for i := 1; i <= 10; i++ {
		msg := fmt.Sprintf("message-%d", i)
		err := writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(fmt.Sprintf("key-%d", i%3)),
				Value: []byte(msg),
			},
		)
		if err != nil {
			log.Println("Failed to write message:", err)
		} else {
			log.Println("Produced:", msg)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// 消费者函数
func consume(groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		GroupID: groupID,
		Topic:   topic,
	})
	defer r.Close()

	log.Printf("Consumer group %s started\n", groupID)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Read error:", err)
			continue
		}
		log.Printf("[Group %s] Consumed: key=%s value=%s partition=%d offset=%d\n",
			groupID, string(m.Key), string(m.Value), m.Partition, m.Offset)
	}
}

func main() {
	// 创建 Topic
	createTopic()

	// 启动生产者
	go produce()

	// 启动两个不同消费组
	go consume("group-A")
	go consume("group-B")

	select {} // 防止主线程退出
}
