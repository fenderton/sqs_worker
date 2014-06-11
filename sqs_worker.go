package main

import "os"
import "sync"
import "flag"
import "fmt"
import "strconv"

import "./heartbeat"
import "./work_order"

import "github.com/crowdmob/goamz/sqs"
import "github.com/ianneub/logger"

const (
  VERSION = "1.1.0"
)

func init() {
  print_version := flag.Bool("v", false, "display version and exit")
  debug := flag.Bool("d", false, "enable debug mode")

  // parse command line options
  flag.Parse()

  // display version and exit
  if *print_version {
    fmt.Println("SQS worker version:", VERSION)
    os.Exit(0)
  }

  // set debug
  if *debug {
    logger.SetDebug(true)
  }
}

func main() {
  // get worker count
  workers, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
  if err != nil {
    workers = 10
  }

  // access key, secret key, receive queue and report queue should be in ENV variables
  logger.Debug("Starting SQS worker version: %s with %d workers.", VERSION, workers)

  // create sqs client
  client, err := sqs.NewFrom(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "us.east")
  if err != nil {
    logger.Fatal("CLIENT ERROR: %v", err)
  }

  // get the SQS queue
  queue, err := client.GetQueue(os.Getenv("SQS_RECIEVE_QUEUE"))
  if err != nil {
    logger.Fatal("QUEUE ERROR: %v", err)
  }

  // get some messages from the sqs queue
  resp, err := queue.ReceiveMessageWithVisibilityTimeout(workers, 60)
  if err != nil {
    logger.Fatal("Could not receive messages: %v", err)
  }

  if cap(resp.Messages) == 0 {
    logger.Debug("Did not find any messages in the queue.")
  }

  // create the wait group
  var wg sync.WaitGroup
  
  // for each message
  for _, message := range resp.Messages {
    // get the message details
    wo, err := work_order.NewFromJson(message.Body)
    if err != nil {
      logger.Error("Could not process SQS message: %s with JSON ERROR: %v", message.MessageId, err)
    } else {
      wg.Add(1)
      go process(queue, message, wo, &wg)
    }
  }

  // wait for each goroutine to exit
  wg.Wait()

  // quit
  logger.Debug("Exiting.")
  os.Exit(0)

}

// process a message from the SQS queue. This should be run inside a goroutine.
func process(q *sqs.Queue, m sqs.Message, wo work_order.WorkOrder, wg *sync.WaitGroup) {
  logger.Debug("Starting process on %d from '%s'", wo.Id, m.MessageId)

  // start heartbeat
  beat := heartbeat.Start(q, &m)
  
  // execute the work
  err := wo.Execute()
  if err != nil {
    logger.Error("Error executing: %d - %v", wo.Id, err)
  }

  // send response back to devops-web
  wo.Report()

  // stop the heartbeat
  beat.Stop()

  // delete message
  logger.Debug("Deleting message: %s", m.MessageId)
  _, err = q.DeleteMessage(&m)
  if err != nil {
    logger.Error("ERROR: Couldn't delete message: %s - %v", m.MessageId, err)
  }

  // exit this goroutine
  wg.Done()
}
