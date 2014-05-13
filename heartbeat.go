package heartbeat

// import "fmt"
import "log"
import "os"
import "os/exec"
import "time"
import "syscall"
// import "reflect"
import "github.com/Mistobaan/sqs"

func heartbeat() {
  // https://gobyexample.com/tickers
  // Setup heartbeat to run in the background and respond to the ticker
  heartbeat := time.NewTicker(4 * time.Minute)
  go func() {
    for t := range heartbeat.C {
      // update SQS with each tick from the heartbeat
      updateSQS(t)
    }
  }()

  // setup command to be run with arguments from the command line
  cmd := exec.Command(os.Args[1])
  cmd.Args = os.Args[1:]
  
  // wire up stdout and stderr with this processes' stdout
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  // execute the command
  if err := cmd.Start(); err != nil {
    log.Fatalf("cmd.Start: %v")
  }

  // http://stackoverflow.com/questions/10385551/get-exit-code-go
  if err := cmd.Wait(); err != nil {
    if exiterr, ok := err.(*exec.ExitError); ok {
      // The program has exited with an exit code != 0

      // This works on both Unix and Windows. Although package
      // syscall is generally platform dependent, WaitStatus is
      // defined for both Unix and Windows and in both cases has
      // an ExitStatus() method with the same signature.
      if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
        // log.Printf("Exit Status: %d", status.ExitStatus())
        // quit <- true
        os.Exit(status.ExitStatus())
      }
    } else {
      log.Printf("cmd.Wait: %v", err)
    }
  }
}

func updateSQS(t time.Time) {
  log.Println("Sending heartbeat to SQS")

  // create the SQS client
  client, err := sqs.NewFrom(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "us.east")
  if err != nil {
    log.Println("ERROR:", err)
  }
  
  // fake an existing sqs.Message by creating it and setting the ReceiptHandle
  message := sqs.Message{}
  message.ReceiptHandle = os.Getenv("RECEIPT_HANDLE")

  // get the SQS queue
  queue, err := client.GetQueue(os.Getenv("SQS_QUEUE_NAME"))
  if err != nil {
    log.Println("ERROR:", err)
  }

  // change the sqs message visibility
  _, err = queue.ChangeMessageVisibility(&message, 5 * 60)
  if (err != nil) {
    log.Println("ERROR:", err)
  }
}
