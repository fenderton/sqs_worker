package work_order

import "os"
import "encoding/json"
import "time"
import "os/exec"
import "syscall"
import "log"
import "bytes"
import "strings"
import "github.com/Mistobaan/sqs"

/*
Example message:

{
  "id": 26487,
  "job_id": 11,
  "completed_at": null,
  "message": "./fixtures/test.sh 0 10",
  "result": "Stock Status index was rebuilt successfully",
  "created_at": "2014-05-07T09:42:11.000-07:00",
  "updated_at": "2014-05-07T09:43:06.000-07:00",
  "exit_status": 0,
  "queue": "sexydresses-worker"
}
*/

type WorkOrder struct {
  Id int `json:"id"`
  JobId int `json:"job_id"`
  CompletedAt *time.Time `json:"completed_at"`
  Message string `json:"message"`
  Result string `json:"result"`
  CreatedAt *time.Time `json:"created_at"`
  UpdatedAt *time.Time `json:"updated_at"`
  ExitStatus int `json:"exit_status"`
  Queue string `json:"queue"`

  response Response
}

type Response struct {
  Id int `json:"id"`
  Result Result `json:"result"`
  CompletedAt *time.Time `json:"completed_at"`
  TimeTaken float64 `json:"time_taken"`
}

type Result struct {
  ExitStatus int `json:"exit_status"`
  Message string `json:"message"`
}

func NewFromJson(data string) (wo WorkOrder, error error){
  bytes := []byte(data)
  err := json.Unmarshal(bytes, &wo)
  if err != nil {
    error = err
  }

  return
}

func (wo *WorkOrder) Execute() (error error) {
  log.Println("Starting work on WorkOrder:", wo.Id)

  // setup command to be run with arguments from the command line
  shell := strings.Split(wo.Message, " ")
  cmd := exec.Command(os.Getenv("CMD_BASE"))
  cmd.Args = shell[0:]
  
  // collect stdout and stderr
  var output bytes.Buffer
  cmd.Stdout = &output
  cmd.Stderr = &output

  // start timing command
  start_time := time.Now()

  // execute the command
  if err := cmd.Start(); err != nil {
    log.Println("cmd.Start:", err)
    error = err
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
        wo.response.Result.ExitStatus = status.ExitStatus()
      }
    } else {
      log.Printf("cmd.Wait: %v", err)
      error = err
    }
  }

  // calculate the time taken to complete the command
  wo.response.TimeTaken = time.Since(start_time).Seconds()

  wo.response.Result.Message = output.String()
  current_time := time.Now()
  wo.response.CompletedAt = &current_time

  log.Println("Completed WorkOrder:", wo.Id)
  return
}

func (wo *WorkOrder) Report() (error error) {
  log.Println("Sending response to devops-web for:", wo.Id)

  // prepare the response object
  wo.response.Id = wo.Id

  // create sqs client
  client, err := sqs.NewFrom(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "us.east")
  if err != nil {
    log.Println("Could not report:", wo.Id, err)
    error = err
  }

  // get the SQS queue
  queue, err := client.GetQueue(os.Getenv("SQS_REPORT_QUEUE"))
  if err != nil {
    log.Println("REPORT QUEUE ERROR:", wo.Id, err)
    error = err
  }

  // marshal the response object into json
  data, err := json.Marshal(wo.response)
  if err != nil {
    log.Println("Could not convert response to JSON for:", wo.Id, err)
    error = err
  }

  // send the report to the queue
  _, err = queue.SendMessage(string(data))
  if err != nil {
    log.Println("Could not report:", wo.Id, err)
    error = err
  }

  return
}
