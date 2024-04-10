package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func producer(quantity int) {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n", os.Args[0])
		os.Exit(1)
	}
	configFile := os.Args[1]
	conf := ReadConfig(configFile)

	// Adjusting the producer configuration for higher performance
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Producer.Flush.Frequency = 500                 // Flush batches every 500ms
	conf.Producer.Flush.MaxMessages = 10000             // Flush when we have 10,000 messages ready to send
	conf.Producer.RequiredAcks = kafka.WaitForAll       // Wait for full acknowledgement
	conf.Producer.Retry.Max = 5                         // Retry up to 5 times
	conf.Producer.Compression = kafka.CompressionSnappy // Use snappy compression

	topic := "topic-khai"
	p, err := kafka.NewProducer(&conf)
	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

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
	p.Flush(15 * 1000)
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

func main() {
	router := gin.Default()

	// Define a GET request handler at '/'
	router.GET("/", handler)

	// Start the server on port 8080
	router.Run(":8089")
}
