package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
)

// JSONMessenger is the interface for sending JSON messages.
type JSONMessenger interface {
	Send(m JSONMessage) error
}

// JSONSqsMessenger is the AWS SQS implementation for the JSON messenger.
type JSONSqsMessenger struct {
	sqsService *sqs.SQS
	queueURL   string
}

// Send sends a message to the queue.
func (q *JSONSqsMessenger) Send(m JSONMessage) error {
	s, err := m.JSON()
	if err != nil {
		return err
	}
	_, err = q.sqsService.SendMessage(&sqs.SendMessageInput{
		MessageBody: &s,
		QueueUrl:    &q.queueURL,
	})
	if err != nil {
		return err
	}
	return nil
}

// NewJSONSqsMessenger creates the new AWS SQS implementation of JSONMessenger.
func NewJSONSqsMessenger(sqsService *sqs.SQS, queueURL string) JSONMessenger {
	return &JSONSqsMessenger{
		sqsService: sqsService,
		queueURL:   queueURL,
	}
}

// NewJSONSqsMessengerFromEnv creates the new AWS SQS implementation of JSONMessenger
// from env vars.
func NewJSONSqsMessengerFromEnv() JSONMessenger {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("CC_AWS_REGION")),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     os.Getenv("CC_AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("CC_AWS_SECRET_ACCESS_KEY"),
		}),
	}))
	return &JSONSqsMessenger{
		sqsService: sqs.New(sess),
		queueURL:   os.Getenv("CC_AWS_SQS_QUEUE"),
	}
}
