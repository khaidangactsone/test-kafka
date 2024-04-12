package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

var producer sarama.SyncProducer

func producerHandler(c *gin.Context) {
	startTime := time.Now()
	quantityString := c.Query("quantity")
	quantity, err := strconv.Atoi(quantityString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quantity"})
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	topic := "test_toppic_khai"
	var messages []*sarama.ProducerMessage

	for i := 0; i < quantity; i++ {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(fmt.Sprintf("Message %d", i)),
		}
		// partition, offset, err := producer.SendMessage(msg)
		messages = append(messages, msg)
		// time.Sleep(1 * time.Second)
	}

	err = producer.SendMessages(messages)
	if err != nil {
		producerErrors, ok := err.(sarama.ProducerErrors)
		if ok {
			for _, err := range producerErrors {
				fmt.Printf("Failed to send message: %s\n", err.Err)
			}
		} else {
			fmt.Printf("Failed to send messages: %s\n", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send messages"})
		return
	} else {
		fmt.Printf("%d messages sent to topic %s\n", quantity, topic)
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	elapsedTime = elapsedTime / time.Millisecond
	c.JSON(http.StatusOK, gin.H{"message": elapsedTime})
}

func producerHasDataHandler(c *gin.Context) {
	// startTime := time.Now()
	var payload interface{}
	c.BindJSON(&payload)
	fmt.Println(payload)
	payloadByte, _ := json.Marshal(payload)
	var message sarama.ProducerMessage
	message.Topic = "test_toppic_khai"
	message.Value = sarama.StringEncoder(payloadByte)

	partition, offset, err := producer.SendMessage(&message)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send messages"})
	}
	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", "test_toppic_khai", partition, offset)
	c.JSON(http.StatusOK, gin.H{"message": payloadByte})
}

type ConsumerGroupHandler struct{}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message claimed: value = %s, timestamp = %v, topic = %s\n", string(msg.Value), msg.Timestamp, msg.Topic)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func consumerHandler(c *gin.Context) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0 // Specify appropriate Kafka version
	config.Consumer.Return.Errors = true
	group, err := sarama.NewConsumerGroup([]string{"192.168.2.45:9092"}, "test_group", config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := group.Close(); err != nil {
			panic(err)
		}
	}()

	consumer := ConsumerGroupHandler{}

	ctx := context.Background()
	topics := []string{"test_toppic_khai"}

	stopChan := make(chan struct{})
	go func() {
		for {
			if err := group.Consume(ctx, topics, consumer); err != nil {
				panic(err)
			}
			select {
			case <-ctx.Done():
				return
			case <-stopChan:
				return

			default:
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the consumer
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	go func() {
		time.Sleep(3 * time.Second)
		stopChan <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		fmt.Println("terminating: context cancelled")
	case <-sigterm:
		fmt.Println("terminating: via signal")
	case <-stopChan:
		fmt.Println("terminating: via signal")
		break
	}
	defer func() {
		fmt.Println("Consumer is stopped")
		close(stopChan)
		close(sigterm)
		ctx.Done()
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Consumer is stopped"})
}

func init() {
	var err error
	producer, err = sarama.NewSyncProducer([]string{"192.168.2.45:9092"}, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	router := gin.Default()

	// Define a GET request handler at '/'
	router.GET("/producer", producerHandler)
	router.GET("/consumer", consumerHandler)
	router.POST("/producer", producerHasDataHandler)

	// go consumerHandler()

	// Start the server on port 8080
	router.Run(":8090")
}
