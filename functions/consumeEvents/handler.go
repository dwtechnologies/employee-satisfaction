/*
== ENV VARS ==
	SQS_URL

*/

package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type server struct {
	svc      *sqs.SQS
	messages []sqs.Message
	delete   []sqs.DeleteMessageBatchRequestEntry

	query string
}

var (
	sqsURL = os.Getenv("SQS_URL")

	host      = os.Getenv("REDSHIFT_HOST")
	port      = os.Getenv("REDSHIFT_PORT")
	db        = os.Getenv("REDSHIFT_DB")
	tableName = os.Getenv("REDSHIFT_TABLE_NAME")
	user      = os.Getenv("REDSHIFT_USERNAME")
	pass      = ""

	clickStates = map[string]int{
		"SINGLE": 1,
		"DOUBLE": 1, // We will treat DOUBLE same as SINGLE.
		"LONG":   3,
	}
)

func main() {
	lambda.Start(Handler)
}

// Handler will check the SQS queue defined by SQS_URL env var for
// new messages. Consuming up to over 509 messages at a time and
// inserts them to the Redshift table defined by the tableName env var.
// Returns string and error.
func Handler() (string, error) {
	if err := checkEnv(); err != nil {
		return "", err
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}
	srv := server{svc: sqs.New(cfg)}

	// Get up to 509 messages from the SQS queue and store them
	// in srv.messages.
	if err := srv.get(); err != nil {
		return "", err
	}

	// Return an empty success message if the queue didn't contain
	// any new messages.
	if len(srv.messages) == 0 {
		return "No new messages on queue", nil
	}

	// Parse messages and create the Redshift SQL Query.
	// Prepare the batch delete at srv.delete.
	if err := srv.parse(); err != nil {
		return "", err
	}

	// Save the results to Redshift cluster.
	if err := srv.save(); err != nil {
		return "", err
	}

	// Delete all messages contained in src.delete.
	if err := srv.del(); err != nil {
		return "", err
	}

	return "Messages pushed to Redshift", nil
}

// checkEnv will check that all the necessary env vars are set.
// It will also decrypt the password to the redshift cluster.
// Returns error.
func checkEnv() error {
	switch {
	case sqsURL == "":
		return fmt.Errorf("Environment variable SQS_URL cannot be empty")

	case host == "":
		return fmt.Errorf("Environment variable REDSHIFT_HOST")

	case port == "":
		return fmt.Errorf("Environment variable REDSHIFT_PORT cannot be empty")

	case db == "":
		return fmt.Errorf("Environment variable REDSHIFT_DB cannot be empty")

	case tableName == "":
		return fmt.Errorf("Environment variable REDSHIFT_TABLE_NAME cannot be empty")

	case user == "":
		return fmt.Errorf("Environment variable REDSHIFT_USERNAME cannot be empty")

	case os.Getenv("REDSHIFT_PASSWORD") == "":
		return fmt.Errorf("Environment variable REDSHIFT_PASSWORD cannot be empty")
	}

	// Decrypt the redshift password.
	return decryptPassword(os.Getenv("REDSHIFT_PASSWORD"))
}
