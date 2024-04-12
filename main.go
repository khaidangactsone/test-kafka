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
var topic = "many_partitions"
var consummerGroup = "group_many_partitions"

var group sarama.ConsumerGroup

var stopChan = make(chan struct{})
var ctx = context.Background()

func producerHandler(c *gin.Context) {
	startTime := time.Now()
	quantityString := c.Query("quantity")
	quantity, err := strconv.Atoi(quantityString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quantity"})
		return
	}

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

func producerOneHandler(c *gin.Context) {
	startTime := time.Now()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(fmt.Sprintf("Message %d", 1)),
	}

	_, _, err := producer.SendMessage(msg)
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
		fmt.Printf("%d messages sent to topic %s\n", 1, topic)
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
	message.Topic = topic
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
	stopChan = make(chan struct{})
	consumer := ConsumerGroupHandler{}

	ctx = context.Background()
	topics := []string{topic}

	go func() {
		for {
			if err := group.Consume(ctx, topics, consumer); err != nil {
				panic(err)
			}
			select {
			case <-ctx.Done():
				fmt.Println("----------Done")
				return
			case <-stopChan:
				fmt.Println("----------")
				return
			default:
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the consumer
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
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

	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0 // Specify appropriate Kafka version
	config.Consumer.Return.Errors = true
	group, err = sarama.NewConsumerGroup([]string{"192.168.2.45:9092"}, consummerGroup, config)
	if err != nil {
		panic(err)
	}
}

func consumerStopHandler(c *gin.Context) {
	if stopChan != nil {
		close(stopChan)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Consumer is stopped"})
}

func main() {
	router := gin.Default()

	// Define a GET request handler at '/'
	router.GET("/producer-only", producerOneHandler)
	router.GET("/producer", producerHandler)
	router.GET("/consumer", consumerHandler)
	router.GET("/consumer-stop", consumerStopHandler)
	router.POST("/producer", producerHasDataHandler)

	// go consumerHandler()

	// Start the server on port 8080
	router.Run(":8090")
}
