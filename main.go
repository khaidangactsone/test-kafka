package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

func ReadConfig(configFile string) kafka.ConfigMap {

	m := make(map[string]kafka.ConfigValue)

	file, err := os.Open(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %s", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			before, after, found := strings.Cut(line, "=")
			if found {
				parameter := strings.TrimSpace(before)
				value := strings.TrimSpace(after)
				m[parameter] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Failed to read file: %s", err)
		os.Exit(1)
	}

	return m

}

func pruducer(quantity int) {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n",
			os.Args[0])
		os.Exit(1)
	}
	configFile := os.Args[1]
	conf := ReadConfig(configFile)

	topic := "toppic-khai"
	p, err := kafka.NewProducer(&conf)

	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	// go func() {
	// 	for e := range p.Events() {
	// 		switch ev := e.(type) {
	// 		case *kafka.Message:
	// 			if ev.TopicPartition.Error != nil {
	// 				fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
	// 			} else {
	// 				fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
	// 					*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
	// 			}
	// 		}
	// 	}
	// }()

	for n := 0; n < quantity; n++ {
		key := "order"
		data := n
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            []byte(key),
			Value:          []byte(fmt.Sprintf("%d", data)),
		}, nil)
	}

	// Wait for all messages to be delivered
	p.Flush(100)
	p.Close()
}

func handler(c *gin.Context) {
	startTime := time.Now()
	quantityString := c.Query("quantity")
	quantity, err := strconv.Atoi(quantityString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quantity"})
		return
	}
	pruducer(quantity)
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	// convert to milliseconds
	elapsedTime = elapsedTime / time.Millisecond
	c.JSON(http.StatusOK, gin.H{"message": elapsedTime})
}

func consumerHandler(con *gin.Context) {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n",
			os.Args[0])
		os.Exit(1)
	}

	configFile := os.Args[1]
	conf := ReadConfig(configFile)
	conf["group.id"] = "test-group-comsumer"
	conf["auto.offset.reset"] = "earliest"

	c, err := kafka.NewConsumer(&conf)

	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
		os.Exit(1)
	}

	topic := "toppic-khai"
	err = c.SubscribeTopics([]string{topic}, nil)
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Errors are informational and automatically handled by the consumer
				continue
			}
			fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
				*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
		}
	}

	c.Close()
	con.JSON(http.StatusOK, gin.H{"message": "Consumer is closed"})
}

func main() {
	router := gin.Default()

	// Define a GET request handler at '/'
	router.GET("/", handler)

	router.GET("/consumer", consumerHandler)

	// Start the server on port 8080
	router.Run(":8090")
}
