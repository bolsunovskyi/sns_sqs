package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	awsKey := flag.String("ak", "", "aws api key")
	awsSecret := flag.String("as", "", "aws api secret")
	awsRegion := flag.String("ar", "us-east-2", "aws region")
	queueUrl := flag.String("q", "", "sqs queue name")
	timeout := flag.Int64("qt", 5, "(Optional) Timeout in seconds for long polling")
	flag.Parse()

	sqsClient := sqs.New(session.Must(session.NewSession(&aws.Config{
		Region: awsRegion,
		Credentials: credentials.NewStaticCredentials(
			*awsKey,
			*awsSecret,
			"",
		),
	})))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	log.Println("starting listen to queue", *queueUrl)
	for {
		select {
		case <-c:
			log.Println("exit")
			os.Exit(0)
		default:
			res, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            queueUrl,
				MaxNumberOfMessages: aws.Int64(10),
				WaitTimeSeconds:     timeout,
			})
			if err != nil {
				log.Fatalln(err)
			}

			for _, m := range res.Messages {
				log.Println(*m.MessageId, *m.Body)

				if _, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      queueUrl,
					ReceiptHandle: m.ReceiptHandle,
				}); err != nil {
					log.Fatalln(err)
				}
			}

		}
	}
}
