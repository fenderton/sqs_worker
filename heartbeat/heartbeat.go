package heartbeat

import "time"
import "github.com/crowdmob/goamz/sqs"
import "github.com/ianneub/logger"

// simple struct that holds a reference to the ticker
type Heartbeat struct {
  ticker *time.Ticker
  Message *sqs.Message
  Queue *sqs.Queue
}

// Start a heartbeat against a given queue and message
func Start(q *sqs.Queue, m *sqs.Message) (heartbeat Heartbeat) {
  logger.Debug("Starting heartbeat on: %s", m.MessageId)

  heartbeat.ticker = time.NewTicker(50 * time.Second)
  heartbeat.Message = m
  heartbeat.Queue = q

  go func() {
    for t := range heartbeat.ticker.C {
      // update SQS with each tick from the heartbeat
      heartbeat.beat(t)
    }
  }()

  return
}

// Stop the heartbeat
func (heartbeat *Heartbeat) Stop() {
  logger.Debug("Stopping heartbeat on: %s", heartbeat.Message.MessageId)
  heartbeat.ticker.Stop()
}

// Send a heartbeat to SQS notifying it that we are still working on the message.
func (heartbeat *Heartbeat) beat(t time.Time) {
  logger.Debug("Sending heartbeat for: %s", heartbeat.Message.MessageId)

  // change the sqs message visibility
  _, err := heartbeat.Queue.ChangeMessageVisibility(heartbeat.Message, 2 * 60)
  if err != nil {
    logger.Error("HEARTBEAT ERROR: messageId: %s - %v", heartbeat.Message.MessageId, err)
  }
}
