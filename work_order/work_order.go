package work_order

import "os"
import "encoding/json"
import "time"
import "os/exec"
import "syscall"
import "bytes"
import "strings"
import "github.com/crowdmob/goamz/sqs"
import "github.com/ianneub/logger"

/*
Example message:

{
  "id": 26487,
  "job_id": 11,
  "completed_at": null,
  "message": "0 10",
  "result": "Stock Status index was rebuilt successfully",
  "created_at": "2014-05-07T09:42:11.000-07:00",
  "updated_at": "2014-05-07T09:43:06.000-07:00",
  "exit_status": 0,
  "queue": "sexydresses-worker"
}
*/

// WorkOrder contains all the details needed to execute and report on a WorkOrder object.
// This struct maps directly to the WorkOrder in the Jobs system.
type WorkOrder struct {
  Id int `json:"id"`
  JobId int `json:"job_id"`
  CompletedAt *time.Time `json:"completed_at"`
  Message string `json:"message"` // command to be executed
  Result string `json:"result"` // both std and err output of the command
  CreatedAt *time.Time `json:"created_at"`
  UpdatedAt *time.Time `json:"updated_at"`
  ExitStatus int `json:"exit_status"` // exit code from the program
  Queue string `json:"queue"` // name of the SQS queue this message was posted to

  response Response
}

// Response object is what the Jobs system looks for the in the completed jobs SQS queue.
// It should contain the result and timing of the WorkOrder's execution.
type Response struct {
  Id int `json:"id"`
  Result Result `json:"result"`
  CompletedAt *time.Time `json:"completed_at"`
  TimeTaken float64 `json:"time_taken"` // time taken to complete execution in seconds
}

// Result contains the exit status and the output of the command that was run.
type Result struct {
  ExitStatus int `json:"exit_status"` // exit code from the program
  Message string `json:"message"` // both std and err output of the command
}

// Create a new WorkOrder struct from a JSON encoded string
func NewFromJson(data string) (wo WorkOrder, error error){
  bytes := []byte(data)
  err := json.Unmarshal(bytes, &wo)
  if err != nil {
    error = err
  }

  return
}

// Execute a WorkOrder and populate its response object
func (wo *WorkOrder) Execute() (error error) {
  logger.Info("Starting work on WorkOrder: %d with \"%s\"", wo.Id, wo.Message)

  // setup command to be run with arguments from the command line
  wo_args := strings.Split(wo.Message, " ")
  base_args := strings.Split(os.Getenv("CMD_BASE"), " ")
  cmd := exec.Command(base_args[0])
  cmd.Args = append(base_args[0:], wo_args[0:]...)
  cmd.Dir = os.Getenv("CMD_DIR")
  
  // collect stdout and stderr
  var output bytes.Buffer
  cmd.Stdout = &output
  cmd.Stderr = &output

  // start timing command
  start_time := time.Now()

  // execute the command
  if err := cmd.Start(); err != nil {
    logger.Error("cmd.Start:", err)
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
        // logger.Info("Exit Status: %d", status.ExitStatus())
        wo.response.Result.ExitStatus = status.ExitStatus()
      }
    } else {
      logger.Error("cmd.Wait: %v", err)
      error = err
    }
  }

  // calculate the time taken to complete the command
  end_time := time.Now()
  wo.response.TimeTaken = end_time.Sub(start_time).Seconds()
  wo.response.CompletedAt = &end_time

  // attach the output of the command to the result message
  wo.response.Result.Message = output.String()  

  logger.Info("Completed WorkOrder: %d", wo.Id)
  return
}

// Report on the result of the WorkOrders execution.
// This method requires that the WorkOrder has been Executed.
func (wo *WorkOrder) Report() (error error) {
  report_queue := os.Getenv("SQS_REPORT_QUEUE")
  logger.Info("Sending response to '%s' for: %d", report_queue, wo.Id)

  // prepare the response object
  wo.response.Id = wo.Id

  // create sqs client
  client, err := sqs.NewFrom(os.Getenv("SQS_WORKER_ACCESS_KEY"), os.Getenv("SQS_WORKER_SECRET_KEY"), "us-east-1")
  if err != nil {
    logger.Error("Could not report: %d - %v", wo.Id, err)
    error = err
    return
  }

  // get the SQS queue
  queue, err := client.GetQueue(report_queue)
  if err != nil {
    logger.Error("REPORT QUEUE ERROR: %d - %v", wo.Id, err)
    error = err
    return
  }

  // trim the message so that it will fit in SQS
  if len(wo.response.Result.Message) > 60000 {
    logger.Info("Trimming WorkOrder %d message down from %d chars", wo.Id, len(wo.response.Result.Message))
    wo.response.Result.Message = wo.response.Result.Message[0:30000] + "\n\n...(truncated)...\n\n" + wo.response.Result.Message[len(wo.response.Result.Message)-30000:]
  }

  // marshal the response object into json
  data, err := json.Marshal(wo.response)
  if err != nil {
    logger.Error("Could not convert response to JSON for: %d - %v", wo.Id, err)
    error = err
    return
  }

  // send the report to the queue
  _, err = queue.SendMessage(string(data))
  if err != nil {
    logger.Error("Could not report: %d - %v", wo.Id, err)
    error = err
  }

  return
}
