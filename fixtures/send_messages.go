package main

import "os"
import "time"
import "strconv"
import "encoding/json"
import "math/rand"

import "../work_order"

import "github.com/crowdmob/goamz/sqs"
import "github.com/ianneub/logger"

func main() {
  logger.Info("Sending messages...")
  
  // create sqs client
  client, err := sqs.NewFrom(os.Getenv("ADMIN_AWS_ACCESS_KEY_ID"), os.Getenv("ADMIN_AWS_SECRET_ACCESS_KEY"), "us-east-1")
  if err != nil {
    logger.Fatal("CLIENT ERROR:", err, "asdf", "asdfdasfd")
  }

  // get the SQS queue
  queue, err := client.GetQueue(os.Getenv("SQS_RECIEVE_QUEUE"))
  if err != nil {
    logger.Fatal("QUEUE ERROR:", err)
  }

  for i := 0; i < 100; i++ {
    var wo work_order.WorkOrder
    current_time := time.Now()

    wo.Id = i
    wo.JobId = 1
    wo.Message = strconv.Itoa(rand.Intn(80))
    wo.CreatedAt = &current_time
    wo.UpdatedAt = &current_time
    wo.Queue = os.Getenv("SQS_RECIEVE_QUEUE")

    data, err := json.Marshal(wo)
    if err != nil {
      logger.Error("JSON error: %v", err)
    }
    queue.SendMessage(string(data))
  }

  // quit
  logger.Info("Exiting.")
  os.Exit(0)

}
