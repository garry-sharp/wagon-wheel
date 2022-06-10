package logger

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type LogLevel uint8

const (
	Info LogLevel = iota
	Warning
	Panic
	Fatal
)

// A Bit mask implementation for defining which logs are wanted

type LogType uint8

const (
	Stdout LogType = 1 << 0
	File   LogType = 1 << 1
	Kafka  LogType = 1 << 2
)

type Config struct {
	LogFilePath string
	WriteType   uint8
	KafkaURL    string

	writeStdout bool
	writeFile   bool
	writeKafka  bool

	logFile     *os.File
	kafkaWriter *kafka.Writer
}

type Topic string

const (
	Price Topic = "Price"
)

var stdoutLogger *log.Logger
var fileLogger *log.Logger
var kafkaLogger *kafka.Writer

var config Config

func Init(c Config) {

	if c.WriteType&uint8(Stdout) == uint8(Stdout) {
		c.writeStdout = true
	}
	if c.WriteType&uint8(File) == uint8(File) {
		c.writeFile = true
		if f, err := os.OpenFile(c.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755); err == nil {
			c.logFile = f
		} else {
			log.Fatal(err)
		}
	}
	if c.WriteType&uint8(Kafka) == uint8(Kafka) {
		c.writeKafka = true
		if c.KafkaURL == "" {
			log.Fatal("Must have a kafka url supplied")
		}

		//Due to library contraints I need to create topics first
		for _, topic := range []Topic{Price} {
			conn, err := kafka.DialLeader(context.Background(), "tcp", c.KafkaURL, string(topic), 0)
			conn.Close()
			if err != nil {
				log.Panic(err.Error())
			}
		}

		c.kafkaWriter = &kafka.Writer{
			Addr:     kafka.TCP(c.KafkaURL),
			Balancer: &kafka.LeastBytes{},
		}

	}
	config = c
}

func Destroy() {
	if config.logFile != nil {
		config.logFile.Close()
	}
	if config.kafkaWriter != nil {
		config.kafkaWriter.Close()
	}
}

func Log(message string, level LogLevel, topic Topic) {
	if config.writeStdout {
		writeStdout(message, level)
	}
	if config.writeFile {
		writeFile(message, level)
	}
	if config.writeKafka {
		writeKafka(message, level, topic)
	}
}

func writeStdout(message string, level LogLevel) {
	if stdoutLogger == nil {
		stdoutLogger = log.New(os.Stdout, "", 0)
	}
	writeLog(message, level, stdoutLogger)
}

func writeFile(message string, level LogLevel) {
	if fileLogger == nil {
		writer := io.Writer(config.logFile)
		fileLogger = log.New(writer, "", 0)
	}
	writeLog(message, level, fileLogger)
}

func writeLog(message string, level LogLevel, logger *log.Logger) {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000000")
	switch level {
	case Fatal:
		logger.Fatalln(currentTime, " Fatal: ", message)
	case Panic:
		logger.Panicln(currentTime, " Error: ", message)
	case Warning:
		logger.Println(currentTime, " Warning: ", message)
	case Info:
		logger.Println(currentTime, " Info: ", message)
	}
}

func writeKafka(message string, level LogLevel, topic Topic) {
	if err := config.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Topic: string(topic),
			Key:   []byte("Message"),
			Value: []byte(message),
		},
	); err != nil {
		writeStdout(err.Error(), Fatal)
	}
}
