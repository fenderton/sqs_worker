package heartbeat

import "time"
import "github.com/Mistobaan/sqs"
import "github.com/ianneub/logger"

// simple struct that holds a reference to the ticker
type Heartbeat struct {
  ticker *time.Ticker
}

// Start a heartbeat against a given queue and message
func Start(q *sqs.Queue, m sqs.Message) (heartbeat Heartbeat) {
  logger.Debug("Starting heartbeat on:", m.MessageId)

  heartbeat.ticker = time.NewTicker(50 * time.Second)
  go func() {
    for t := range heartbeat.ticker.C {
      // update SQS with each tick from the heartbeat
      beat(q, m, t)
    }
  }()

  return
}

// Stop the heartbeat
func (heartbeat Heartbeat) Stop() {
  heartbeat.ticker.Stop()
}

// Send a heartbeat to SQS notifying it that we are still working on the message.
func beat(queue *sqs.Queue, message sqs.Message, t time.Time) {
  logger.Debug("Sending heartbeat for:", message.MessageId)

  // change the sqs message visibility
  _, err := queue.ChangeMessageVisibility(&message, 2 * 60)
  if (err != nil) {
    logger.Error("HEARTBEAT ERROR:", err)
  }
}
