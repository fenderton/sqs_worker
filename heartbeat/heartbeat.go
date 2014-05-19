package heartbeat

import "time"
import "github.com/crowdmob/goamz/sqs"
import "github.com/ianneub/logger"

// simple struct that holds a reference to the ticker
type Heartbeat struct {
  ticker *time.Ticker
  MessageId string
}

// Start a heartbeat against a given queue and message
func Start(q *sqs.Queue, m sqs.Message) (heartbeat Heartbeat) {
  logger.Debug("Starting heartbeat on: %s", m.MessageId)

  heartbeat.ticker = time.NewTicker(50 * time.Second)
  heartbeat.MessageId = m.MessageId
  go func() {
    for t := range heartbeat.ticker.C {
      // update SQS with each tick from the heartbeat
      beat(q, m, t)
    }
  }()

  return
}

// Stop the heartbeat
func (heartbeat *Heartbeat) Stop() {
  logger.Debug("Stopping heartbeat on: %s", heartbeat.MessageId)
  heartbeat.ticker.Stop()
}

// Send a heartbeat to SQS notifying it that we are still working on the message.
func beat(queue *sqs.Queue, message sqs.Message, t time.Time) {
  logger.Debug("Sending heartbeat for: %s", message.MessageId)

  // change the sqs message visibility
  _, err := queue.ChangeMessageVisibility(&message, 2 * 60)
  if err != nil {
    logger.Error("HEARTBEAT ERROR: messageId: %s %v", message.MessageId, err)
  }
}
