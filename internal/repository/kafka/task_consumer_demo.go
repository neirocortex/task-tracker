package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	kafkaImp "github.com/segmentio/kafka-go"
)

type TaskConsumer struct {
	reader *kafkaImp.Reader
}

func NewTaskConsumer(brokerAddress string, producer *TaskNotyfier) *TaskConsumer {
	if producer == nil {
		return nil
	}

	taskConsumer := &TaskConsumer{
		reader: kafkaImp.NewReader(kafkaImp.ReaderConfig{
			Brokers:  []string{brokerAddress},
			Topic:    "task-events",
			GroupID:  "medical-task-tracker-consumer",
			MinBytes: 10,
			MaxBytes: 10e6,
			MaxWait:  1 * time.Second,
		}),
	}

	go taskConsumer.Start()

	return taskConsumer
}

func (c *TaskConsumer) Start() {
	slog.Info("Kafka background consumer started successfully, listening to 'task-events'...")

	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Kafka consumer encountered an error while reading message", "error", err)

			time.Sleep(2 * time.Second)
			continue
		}

		var event TaskCreateEvent
		err = json.Unmarshal(msg.Value, &event)
		if err != nil {
			slog.Error("Kafka consumer failed to unmarshal JSON payload",
				"error", err,
				"raw_bytes", string(msg.Value),
			)
			continue
		}

		slog.Info("[Kafka Consumer Notification]",
			"event", "TASK_CREATED_ASYNC",
			"task_id", event.ID,
			"title", event.Title,
			"description", event.Description,
			"status", event.Status,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
	}
}

func (c *TaskConsumer) Close() error {
	return c.reader.Close()
}
