package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}
}

// LINE APIのレスポンス用
type Response struct {
	Message string `json:"message"`
}

// 引数で受け取った文字列をLINE通知する
func SendLine(result string) (*http.Response, error) {
	accessToken := os.Getenv("TOKEN")
	msg := result

	URL := os.Getenv("LINE_POST_URL")
	u, err := url.ParseRequestURI(URL)
	if err != nil {
		log.Fatal(err)
	}

	c := &http.Client{}

	form := url.Values{}
	form.Add("message", msg)

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return res, nil
}

// AWSの使用料を返す
func GetBilling() string {

	//Must be in YYYY-MM-DD Format
	start := "2021-01-01"
	end := "2021-02-01"
	granularity := "MONTHLY"
	metrics := []string{
		"UnblendedCost",
	}
	// Initialize a session in us-east-1 that the SDK will use to load credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create Cost Explorer Service Client
	svc := costexplorer.New(sess)

	result, err := svc.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: aws.String(granularity),
		GroupBy: []*costexplorer.GroupDefinition{
			&costexplorer.GroupDefinition{
				Type: aws.String("DIMENSION"),
				Key:  aws.String("SERVICE"),
			},
		},
		Metrics: aws.StringSlice(metrics),
	})
	if err != nil {
		exitErrorf("Unable to generate report, %v", err)
	}
	total := 0.0
	// なぜかTOTALがブランクなので計算して返す
	for _, group := range result.ResultsByTime[0].Groups {
		fee, err := strconv.ParseFloat(*group.Metrics["UnblendedCost"].Amount, 64)
		if err != nil {
			log.Fatal(err)
		}
		total += fee
	}
	// json, _ := json.Marshal(result.ResultsByTime[0])
	return "今月のAWS使用料" + strconv.FormatFloat(total, 'f', -1, 64) + "USD"
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	billing := GetBilling()
	SendLine(billing)
}
