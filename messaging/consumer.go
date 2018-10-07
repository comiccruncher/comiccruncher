package messaging

import (
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
	"os"
	"sync"
)

// A handler for applying a function to an incoming sync message.
type SyncMessageFunc func(message *SyncMessage)

type SyncMessageConsumer struct {
	sqsService          *sqs.SQS
	maxNumberOfMessages int64
	visibilityTimeout   int64
	waitTimeSeconds     int64
	syncMessageHandler  SyncMessageFunc
	queueUrl            string
}

// Sequentially consume messages from the sqs.
func (c *SyncMessageConsumer) Consume(indefinite bool) error {
	result, err := c.receive()
	if err != nil {
		return err
	}
	for _, m := range result.Messages {
		c.handleMessage(m, c.syncMessageHandler, nil)
	}
	log.MESSAGING().Info("Done processing messages.", zap.Int("messages", len(result.Messages)))
	if indefinite {
		c.Consume(indefinite)
	}
	return nil
}

// Consumes messages concurrently by the max number of messages to receive in the struct settings.
/*
func (c *SyncMessageConsumer) ConsumeConcurrent(indefinite bool, concurrencyLimit int) error {
	result, err := c.receive()
	if err != nil {
		return err
	}
	msgLength := len(result.Messages)
	if msgLength > 0 {
		var wg sync.WaitGroup
		wg.Add(msgLength)
		for _, m := range result.Messages {
			go func(wg *sync.WaitGroup, message *sqs.Message) {
				c.handleMessage(message, c.syncMessageHandler, wg)
			}(&wg, m)
		}
		wg.Wait() // wait for the wait groups to be done.
		log.MESSAGING().Info("done concurrently processing messages", zap.Int("messages", msgLength))
	}
	return nil
}
*/

func (c *SyncMessageConsumer) receive() (*sqs.ReceiveMessageOutput, error) {
	result, err := c.sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &c.queueUrl,
		MaxNumberOfMessages: aws.Int64(c.maxNumberOfMessages),
		VisibilityTimeout:   aws.Int64(c.visibilityTimeout), // 10 hours
		WaitTimeSeconds:     aws.Int64(c.waitTimeSeconds),
	})
	if err != nil {
		return result, err
	}
	log.MESSAGING().Info("number of messages to process.", zap.Int("count", len(result.Messages)))
	return result, err
}

func (c *SyncMessageConsumer) handleMessage(m *sqs.Message, handler SyncMessageFunc, wg *sync.WaitGroup) {
	defer c.sqsService.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &c.queueUrl,
		ReceiptHandle: m.ReceiptHandle,
	})
	message, err := NewSyncMessageFromString(*m.Body)
	if err != nil {
		if wg != nil {
			defer wg.Done()
		}
		log.MESSAGING().Error("error unmarshalling message from body", zap.Error(err), zap.String("body", *m.Body))
		return
	}
	if wg != nil {
		// Apply the handler to the sync message.
		go func(message *SyncMessage, wg *sync.WaitGroup) {
			defer wg.Done()
			handler(message)
		}(message, wg)
	} else {
		handler(message)
	}
}

func NewSyncMessageConsumerFromEnv(maxNumberOfMessages int64, waitTimeInSeconds int64, handler SyncMessageFunc) *SyncMessageConsumer {
	creds := credentials.Value{
		AccessKeyID:     os.Getenv("CC_AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("CC_AWS_SECRET_ACCESS_KEY"),
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("CC_AWS_REGION")),
		Credentials: credentials.NewStaticCredentialsFromCreds(creds),
	}))
	sqsService := sqs.New(sess)
	return &SyncMessageConsumer{
		sqsService:          sqsService,
		maxNumberOfMessages: maxNumberOfMessages,
		visibilityTimeout:   3000,
		waitTimeSeconds:     waitTimeInSeconds,
		queueUrl:            os.Getenv("CC_AWS_SQS_QUEUE"),
		syncMessageHandler:  handler,
	}
}
