/*
This lambda function will push serialNumber, clickType and current DateTime
to the SQS queue specified with the SQS_URL env variable.

If a voltage below 1500mV is detected it will send a warning to the
SNS topic specified with the SNS_TOPIC env variable.

== ENV VARS ==
	SQS_URL (URL OF SQS QUEUE TO PUBLISH TO)
	SNS_TOPIC (ARN TO SNS TOPIC FOR VOLTAGE WARNING)

== INPUT ==
	{
		"serialNumber": "TestSerialNumber",
		"clickType": "SINGLE",
		batteryVoltage: "1604mv"
	}

== OUTPUT ==
	null
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
)

// Time Format to be used to the receiving end (eg. redshift etc).
// Used for converting time.Now() to a string the database can use.
const timeFormat = "2006-01-02 15:04:05.000000"

var (
	sqsURL   = os.Getenv("SQS_URL")
	snsTopic = os.Getenv("SNS_TOPIC")
)

type iotButton struct {
	SerialNumber   string `json:"serialNumber"`
	ClickType      string `json:"clickType"`
	BatteryVoltage string `json:"batteryVoltage"`
	dateTime       string
	cfg            aws.Config
}

func main() {
	lambda.Start(Handler)
}

// Handler takes req and will unmarshal it and send it to an SQS queue.
// If voltage of the IoTButton is low it will send an alarm to an SNS topic.
// Returns string and error.
func Handler(req json.RawMessage) (string, error) {
	if err := checkEnv(); err != nil {
		return "", err
	}

	// Load the default AWS config.
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}

	// Unmarshal the json.RawMessage to button and set current dateTime.
	button := &iotButton{
		cfg:      cfg,
		dateTime: time.Now().Format(timeFormat),
	}
	if err := json.Unmarshal(req, button); err != nil {
		return "", err
	}

	// Publish the message to the SQS queue.
	if err := button.publish(); err != nil {
		return "", err
	}

	// Send warning if volate is below 1500mv.
	if err := button.checkBattery(); err != nil {
		return "", err
	}

	return "Message sent to queue", nil
}

// checkEnv will check that all the necessary env vars are set.
// Returns error.
func checkEnv() error {
	switch {
	case sqsURL == "":
		return fmt.Errorf("Environment variable SQS_URL cannot be empty")

	case snsTopic == "":
		return fmt.Errorf("Environment variable SNS_TOPIC cannot be empty")
	}

	return nil
}
