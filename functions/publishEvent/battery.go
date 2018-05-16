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
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// checkBattery will check that the battery has 1500 or more mV in it. Otherwise
// it will send a warning to an SNS Topic.
// Returns error.
func (b *iotButton) checkBattery() error {
	voltage, err := strconv.Atoi(strings.TrimSpace(strings.Replace(strings.ToLower(b.BatteryVoltage), "mv", "", -1)))
	if err != nil {
		return err
	}

	if voltage >= 1500 {
		return nil
	}

	return b.sendVoltageAlert()
}

// sendVoltageAlert will send an low voltage alert to the SNS Topic specified by SNS_TOPIC.
// Returns error.
func (b *iotButton) sendVoltageAlert() error {
	svc := sns.New(b.cfg)

	_, err := svc.PublishRequest(&sns.PublishInput{
		Subject:  aws.String(fmt.Sprintf("WARNING: IoT Button (%s) has low voltage", b.SerialNumber)),
		Message:  aws.String(fmt.Sprintf("Low voltage has been detected on IoT Button (%s).\n\nCurrent volate: %s", b.SerialNumber, b.BatteryVoltage)),
		TopicArn: &snsTopic,
	}).Send()

	return err
}
