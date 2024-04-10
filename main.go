package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// func ReadConfig(configFile string) kafka.ConfigMap {

// 	m := make(map[string]kafka.ConfigValue)

// 	file, err := os.Open(configFile)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Failed to open file: %s", err)
// 		os.Exit(1)
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := strings.TrimSpace(scanner.Text())
// 		if !strings.HasPrefix(line, "#") && len(line) != 0 {
// 			before, after, found := strings.Cut(line, "=")
// 			if found {
// 				parameter := strings.TrimSpace(before)
// 				value := strings.TrimSpace(after)
// 				m[parameter] = value
// 			}
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		fmt.Printf("Failed to read file: %s", err)
// 		os.Exit(1)
// 	}

// 	return m

// }

// func pruducer(name string) {

// 	if len(os.Args) != 2 {
// 		fmt.Fprintf(os.Stderr, "Usage: %s <config-file-path>\n",
// 			os.Args[0])
// 		os.Exit(1)
// 	}
// 	configFile := os.Args[1]
// 	conf := ReadConfig(configFile)

// 	topic := "toppic-khai"
// 	p, err := kafka.NewProducer(&conf)

// 	if err != nil {
// 		fmt.Printf("Failed to create producer: %s", err)
// 		os.Exit(1)
// 	}

// 	// Go-routine to handle message delivery reports and
// 	// possibly other event types (errors, stats, etc)
// 	go func() {
// 		for e := range p.Events() {
// 			switch ev := e.(type) {
// 			case *kafka.Message:
// 				if ev.TopicPartition.Error != nil {
// 					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
// 				} else {
// 					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
// 						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
// 				}
// 			}
// 		}
// 	}()

// 	users := [...]string{name}
// 	items := [...]string{name}

// 	for n := 0; n < 10; n++ {
// 		key := users[rand.Intn(len(users))]
// 		data := items[rand.Intn(len(items))]
// 		p.Produce(&kafka.Message{
// 			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
// 			Key:            []byte(key),
// 			Value:          []byte(data),
// 		}, nil)
// 	}

// 	// Wait for all messages to be delivered
// 	p.Flush(15 * 1000)
// 	p.Close()

// }

func router(c *gin.Context) {
	name := c.Query("name")

	c.JSON(http.StatusOK, gin.H{"message": name})
}

func main() {
	router := gin.Default()

	// Define a GET request handler at '/'
	router.GET("/", pruducer)

	// Start the server on port 8080
	router.Run(":8089")
}
