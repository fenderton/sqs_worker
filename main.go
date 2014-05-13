package main

import "os"
import "log"
import "sync"
import "./heartbeat"
import "./work_order"
import "github.com/Mistobaan/sqs"

func main() {
  // access key, secret key, receive queue and report queue should be in ENV variables
  log.Println("Starting SQS worker...")

  // create sqs client
  client, err := sqs.NewFrom(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "us.east")
  if err != nil {
    log.Fatalf("CLIENT ERROR:", err)
  }

  // get the SQS queue
  queue, err := client.GetQueue(os.Getenv("SQS_RECIEVE_QUEUE"))
  if err != nil {
    log.Fatalf("QUEUE ERROR:", err)
  }

  // get some messages from the sqs queue
  resp, err := queue.ReceiveMessageWithVisibilityTimeout(4, 60)
  if err != nil {
    log.Fatalf("Could not receive messages:", err)
  }

  if cap(resp.Messages) == 0 {
    log.Println("Did not find any messages in the queue.")
  }

  // create the wait group
  var wg sync.WaitGroup
  
  // for each message
  for _, message := range resp.Messages {
    // get the message details
    wo, err := work_order.NewFromJson(message.Body)
    if err != nil {
      log.Println("Could not process SQS message:", message.MessageId)
      log.Println("JSON ERROR:", err)
    } else {
      wg.Add(1)
      go func() {
        // start heartbeat
        heartbeat.Start(queue, message)
        
        // execute the work
        err := wo.Execute()
        if err != nil {
          log.Println("Error executing ", wo.Id, err)
        }

        // send response back to devops-web
        wo.Report()

        // delete message
        log.Println("Deleting message:", message.MessageId)
        queue.DeleteMessage(&message)

        // exit this goroutine
        wg.Done()
      }()
    } // if err
  } // for

  // wait for each goroutine to exit
  wg.Wait()

  // quit
  log.Println("Exiting.")
  os.Exit(0)

}
