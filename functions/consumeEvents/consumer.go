package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/dwtechnologies/namegen"
)

// sqsMessageResult is sent by the msgchan channel
// to return all the messages it has received as
// well as any error.
type result struct {
	messages []sqs.Message
	err      error
}

var (
	maxBatchItems = 10
	numFetchers   = 50
)

// get will fetch start 50 fetchers against the SQS queue and download
// up to 10 messages per fetcher. So a maximum of 500 messages will be
// collected. If all the fetchers fails with an error we will return
// with an error. Otherwise we will just log a warning to CloudWatch Logs
// with how many fetchers failed.
// Returns error.
func (srv *server) get() error {
	reschan := make(chan *result, numFetchers)
	for i := 0; i < numFetchers; i++ {
		go func(reschan chan *result) {
			// Fetch up to 10 messages from SQS.
			res, err := srv.svc.ReceiveMessageRequest(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(int64(maxBatchItems)),
				QueueUrl:            &sqsURL,
				MessageAttributeNames: []string{
					"serialNumber",
					"clickType",
					"dateTime",
				},
			}).Send()

			reschan <- &result{messages: res.Messages, err: err}
		}(reschan)
	}

	// Loop over the channel until it's complete and append all the messages
	// We received that didn't contain any errors. If we got more than 10 errors
	// log a warning to CloudWatch Logs. If we got 50 errors (100%) then return
	// error.
	errorCount := 0
	lastErrorMsg := ""

	for i := 0; i < numFetchers; i++ {
		res := <-reschan

		switch {
		case res.err != nil:
			errorCount++
			lastErrorMsg = res.err.Error()

		default:
			srv.messages = append(srv.messages, res.messages...)
		}
	}

	switch {
	case errorCount == 50:
		return fmt.Errorf("All 50 of the SQS fetchers failed. Last error message %s", lastErrorMsg)
	case errorCount > 10:
		log.Printf("WARNING: %d SQS fetchers failed. Last error message %s", errorCount, lastErrorMsg)
	}

	log.Printf("Consumed %d messages from queue", len(srv.messages))
	return nil
}

// parse will parse messages in srv.messages and create an array with SQL values
// syntax that can be joined into a SQL query. As well as adding the parsed messages
// to srv.delete for batch deletion once the messages have been sent to Redshift.
// Any messages errors such as invalid clickType or ID generation will be logged
// to CloudWatch logs, but the messages will still be deleted from the SQS queue.
// Returns error.
func (srv *server) parse() error {
	messages := []string{}

	for _, msg := range srv.messages {
		// Add any messages (successfull or not) to the src.delete slice.
		srv.delete = append(srv.delete, sqs.DeleteMessageBatchRequestEntry{
			Id:            msg.MessageId,
			ReceiptHandle: msg.ReceiptHandle,
		})

		// Set and/or check values from the Attributes of the message.
		dateTime := *msg.MessageAttributes["dateTime"].StringValue
		serial := *msg.MessageAttributes["serialNumber"].StringValue
		click, ok := clickStates[*msg.MessageAttributes["clickType"].StringValue]
		if !ok {
			log.Printf("WARNING: Unsupported clickType (%s) received from serialNumber (%s)", *msg.MessageAttributes["clickType"].StringValue, serial)
			continue
		}

		// Generate an uuid for the id column in the table.
		// Log to CloudWatch Logs if generating the uuid fails.
		uuid, err := namegen.Generate()
		if err != nil {
			log.Printf("WARNING: Couldn't generate ID for entry. serialNumber (%s) with state (%d)", serial, click)
			continue
		}

		// Append the message in SQL insert values syntax to the messages slice.
		messages = append(messages, fmt.Sprintf("('%s', '%s', '%s', %d)", uuid, dateTime, serial, click))
	}

	// Generate the query to be used by Redshift to insert the data pulled from SQS.
	srv.build(messages)

	return nil
}

// build will build the query containing all the messages and set srv.query with
// the query containing all the inserts to be made.
func (srv *server) build(messages []string) {
	srv.query = fmt.Sprintf("INSERT INTO %s VALUES %s;", tableName, strings.Join(messages, " ,"))
}

// del will send a DeleteMessageBatch Request to delete all the messages stored in srv.delete.
// It will divide all srv.delete into batches of 10, since this is the max batch size sqs supports.
// It will send each request on a separate go routine to maximize concurrency. The function will only
// Return error if all the messages failed to delete.
// Returns error.
func (srv *server) del() error {
	deleteNum := len(srv.delete) / maxBatchItems
	deleteOverFlow := len(srv.delete) % maxBatchItems
	if deleteOverFlow != 0 {
		deleteNum++
	}

	reschan := make(chan error, deleteNum)

	// Divide the number of messages by 10, since thats our largest batch size.
	// And send as many DeleteMessageBatch requests as necessary.
	for i := 0; i < deleteNum; i++ {
		delete := []sqs.DeleteMessageBatchRequestEntry{}

		switch {
		case i == deleteNum-1:
			delete = srv.delete[maxBatchItems*i : deleteOverFlow]

		default:
			delete = srv.delete[maxBatchItems*i : maxBatchItems*(i+1)]
		}

		go func(delete []sqs.DeleteMessageBatchRequestEntry, reschan chan error) {
			_, err := srv.svc.DeleteMessageBatchRequest(&sqs.DeleteMessageBatchInput{
				Entries:  delete,
				QueueUrl: &sqsURL,
			}).Send()

			reschan <- err
		}(delete, reschan)
	}

	// Wait for all results to finish.
	for i := 0; i < deleteNum; i++ {
		err := <-reschan
		if err != nil {
			//TODO: Do something more meaningful with these errors...
			log.Printf("WARNING: Couldn't delete SQS messages. Error %s", err.Error())
		}
	}

	return nil
}
