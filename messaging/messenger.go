package messaging

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
)

type JsonMessenger interface {
	Send(m JsonMessage) error
}

// A messenger for sending and receiving structs into JSON messages.
type JsonSqsMessenger struct {
	sqsService *sqs.SQS
	queueUrl   string
}

// Sends a message to the queue.
func (q *JsonSqsMessenger) Send(m JsonMessage) error {
	s, err := m.Json()
	if err != nil {
		return err
	}
	_, err = q.sqsService.SendMessage(&sqs.SendMessageInput{
		MessageBody: &s,
		QueueUrl:    &q.queueUrl,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewJsonSqsMessenger(sqsService *sqs.SQS, queueUrl string) JsonMessenger {
	return &JsonSqsMessenger{
		sqsService: sqsService,
		queueUrl:   queueUrl,
	}
}

func NewJsonSqsMessengerFromEnv() JsonMessenger {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("CC_AWS_REGION")),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     os.Getenv("CC_AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("CC_AWS_SECRET_ACCESS_KEY"),
		}),
	}))
	return &JsonSqsMessenger{
		sqsService: sqs.New(sess),
		queueUrl:   os.Getenv("CC_AWS_SQS_QUEUE"),
	}
}
