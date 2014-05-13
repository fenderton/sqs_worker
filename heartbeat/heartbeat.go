package heartbeat

import "log"
import "time"
import "github.com/Mistobaan/sqs"

func Start(q *sqs.Queue, m sqs.Message) {
  log.Println("Starting heartbeat on:", m.MessageId)

  heartbeat := time.NewTicker(50 * time.Second)
  go func() {
    for t := range heartbeat.C {
      // update SQS with each tick from the heartbeat
      beat(q, m, t)
    }
  }()
}

func beat(queue *sqs.Queue, message sqs.Message, t time.Time) {
  log.Println("Sending heartbeat for:", message.MessageId)

  // change the sqs message visibility
  _, err := queue.ChangeMessageVisibility(&message, 2 * 60)
  if (err != nil) {
    log.Println("ERROR:", err)
  }
}
