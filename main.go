package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

//func init() {
//	err := godotenv.Load()
//	if err != nil {
//		log.Fatal("Error loading .env")
//	}
//}

// LINE APIのレスポンス用
type Response struct {
	Message string `json:"message"`
}

// 引数で受け取った文字列をLINE通知する
func SendLine(result string) (*http.Response, error) {
	accessToken := os.Getenv("LINEnotyfyToken")
	msg := result

	URL := os.Getenv("LINEpostURL")
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
	year, monthBase, dayBase := time.Now().Date()
	startMonth := fmt.Sprintf("%02d", monthBase-1)
	endMonth := fmt.Sprintf("%02d", monthBase)
	day := fmt.Sprintf("%02d", dayBase)
	start := strconv.Itoa(year) + "-" + startMonth + "-" + day
	if monthBase == 1 {
		start = strconv.Itoa(year) + "-" + "12-" + day
	}
	end := strconv.Itoa(year) + "-" + endMonth + "-" + day
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

func HandleRequest(ctx context.Context) (string, error) {
	billing := GetBilling()
	res, _ := SendLine(billing)
	ctx.Done()
	return res.Status, nil
}

func main() {
	lambda.Start(HandleRequest)
}
