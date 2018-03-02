package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type event struct {
	Step      string `json:"step"`
	SubStep   string `json:"substep"`
	Timestamp int64  `json:"timestamp"`
	Default   string `json:"default"`
}

func main() {
	awsKey := flag.String("ak", "", "aws api key")
	awsSecret := flag.String("as", "", "aws api secret")
	awsRegion := flag.String("ar", "us-east-2", "aws region")
	snsTopicArn := flag.String("st", "", "sns topic arn")
	flag.Parse()

	snsClient := sns.New(session.Must(session.NewSession(&aws.Config{
		Region: awsRegion,
		Credentials: credentials.NewStaticCredentials(
			*awsKey,
			*awsSecret,
			"",
		),
	})))

	rand.Seed(time.Now().Unix())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	stepI := 0

	for {

		select {
		case <-c:
			log.Println("exit")
			os.Exit(0)
		default:
			step := steps[stepI%len(steps)]
			e := event{
				Step:      step,
				Default:   step,
				SubStep:   substeps[rand.Int31n(int32(len(substeps)))],
				Timestamp: time.Now().Unix(),
			}
			stepI++

			bts, err := json.Marshal(e)
			if err != nil {
				log.Fatalln(err)
			}
			log.Printf("%s - %s\n", e.Step, e.SubStep)

			if _, err := snsClient.Publish(&sns.PublishInput{
				Subject:          aws.String(fmt.Sprintf("%s.%s", e.Step, e.SubStep)),
				MessageStructure: aws.String("json"),
				Message:          aws.String(string(bts)),
				TopicArn:         snsTopicArn,
			}); err != nil {
				log.Fatalln(err)
			}

			time.Sleep(time.Second)
		}
	}
}
