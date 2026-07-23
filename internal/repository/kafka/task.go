package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"strconv"
	"taskTracker/internal/domain"

	kafkaImp "github.com/segmentio/kafka-go"
)

type TaskNotyfier struct {
	writer *kafkaImp.Writer
}

func createTopic(brokerAddress, topic string, partitions int) error {
	conn, err := kafkaImp.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafkaImp.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfig := kafkaImp.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		return err
	}

	slog.Info("Successfully created kafka topic manually", "topic", topic, "partitions", partitions)
	return nil
}

func NewTaskNotyfier(brokerAddress string) *TaskNotyfier {
	slog.Info("Starting kafka sender on", "addr", brokerAddress)

	err := createTopic(brokerAddress, "task-events", 3) // 3 партиции для параллельной обработки
	if err != nil {
		slog.Error("Kafka topic initialization failed", "error", err)
		return nil
	}

	return &TaskNotyfier{
		writer: &kafkaImp.Writer{
			Addr:     kafkaImp.TCP(brokerAddress),
			Topic:    "task-events",
			Balancer: &kafkaImp.LeastBytes{},
			Async:    false,
		},
	}
}

func (tn *TaskNotyfier) Close() {
	tn.writer.Close()
}

type TaskCreateEvent struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (tn *TaskNotyfier) SendCreate(ctx context.Context, task *domain.Task) {
	event := &TaskCreateEvent{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshall create task event", "error", err)
		return
	}

	err = tn.writer.WriteMessages(ctx, kafkaImp.Message{
		Key:   strconv.AppendInt(nil, event.ID, 10),
		Value: payload,
	})
	if err != nil {
		slog.Error("Failed to write event", "error", err)
		return
	}

	slog.Info("Successfully published task event to kafka", "task_id", task.ID)
}
