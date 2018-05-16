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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// publish will send the serialNumber, clickType and dateTime to the SQS Queue specified
// by the sqsURL variable. Returns error.
func (b *iotButton) publish() error {
	// Send a message to the sqs queue URL specified by sqsURL.
	svc := sqs.New(b.cfg)

	_, err := svc.SendMessageRequest(&sqs.SendMessageInput{
		MessageAttributes: map[string]sqs.MessageAttributeValue{
			"serialNumber": sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: &b.SerialNumber,
			},
			"clickType": sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: &b.ClickType,
			},
			"dateTime": sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: &b.dateTime,
			},
		},
		MessageBody: aws.String(fmt.Sprintf("Message from (%s) with clickType (%s)", b.SerialNumber, b.ClickType)),
		QueueUrl:    &sqsURL,
	}).Send()

	return err
}
