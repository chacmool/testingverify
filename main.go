package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/peterbourgon/ff"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	openapiverify "github.com/twilio/twilio-go/rest/verify/v2"
)

type conf struct {
	sms             bool
	ver             bool
	chk             int
	accountSid      string
	authToken       string
	verifyServiceId string
	sendTo          string
	sendFrom        string
}

func main() {

	// openapi "github.com/twilio/twilio-go/rest/verify/v2"
	// openapi "github.com/twilio/twilio-go/rest/api/v2010"

	conf := conf{}
	fs := flag.NewFlagSet("testing-twilio", flag.ExitOnError)

	_ = fs.String("config", "config.conf", "config file (optional)")

	fs.BoolVar(&conf.sms, "sms", false, "send regular sms?")
	fs.BoolVar(&conf.ver, "ver", false, "send verification sms?")
	fs.IntVar(&conf.chk, "chk", 0, "verification check number")

	fs.StringVar(&conf.accountSid, "accountSid", "tbd", "twilio id account")
	fs.StringVar(&conf.authToken, "authToken", "tbd", "twilio token")
	fs.StringVar(&conf.verifyServiceId, "verifyServiceId", "tbd", "twilio verify service id")

	fs.StringVar(&conf.sendTo, "sendTo", "tbd", "mobile phone destination")
	fs.StringVar(&conf.sendFrom, "sendFrom", "tbd", "mobile phone origin")

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser))

	// fmt.Println(conf)

	client := twilio.NewRestClient(conf.accountSid, conf.authToken)

	channelSms := make(chan string)
	channelVer := make(chan string)

	if conf.sms {
		go SendSMS(client, conf.sendTo, conf.sendFrom, "Mensaje de prueba de sms normal desde SALTO", channelSms)
	}

	if conf.ver {
		go SendVerificationMsg(client, conf.sendTo, conf.verifyServiceId, channelVer)
	}

	if conf.chk > 0 {
		fmt.Println("checking with code")
		CheckVerificationMsg(client, conf.sendTo, conf.verifyServiceId, strconv.Itoa(conf.chk))
	}

	if !conf.sms && !conf.ver && conf.chk == 0 {
		fmt.Println("nothing to do")

		// rateLimitId := "RK54a57c23889846b4768ca04294f9c802"
		// ListRateLimits(client, conf.verifyServiceId)
		// CreateRateLimit(client, conf.verifyServiceId)
		// CreateBucket(client, conf.verifyServiceId, rateLimitId)

	}

	if conf.sms {
		fmt.Println(<-channelSms)
	}

	if conf.ver {
		fmt.Println(<-channelVer)
	}

}

func SendSMS(client *twilio.RestClient, to string, from string, msg string, c chan string) {

	go fmt.Println("sending sms to", to)

	params := &openapi.CreateMessageParams{}

	params.SetTo(to)
	params.SetFrom(from)
	params.SetBody(msg)

	resp, err := client.ApiV2010.CreateMessage(params)
	if err != nil {
		fmt.Println("Sending error to", to, err.Error())
		c <- "error sending sms!"
	} else {
		fmt.Println("Response: " + *resp.Status + " - " + *resp.To)
		c <- "sended sms!"
	}

}

func SendVerificationMsg(client *twilio.RestClient, to string, servideId string, c chan string) {

	go fmt.Println("sending verification to", to)
	params := &openapiverify.CreateVerificationParams{}

	params.SetTo(to)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification(servideId, params)
	if err != nil {
		fmt.Println(err.Error())
		c <- "error sending ver!"
	} else {
		fmt.Println("Response: " + *resp.Status + " - " + *resp.To)
		c <- "sended ver!"
	}

}

func CheckVerificationMsg(client *twilio.RestClient, to string, servideId string,
	code string) {

	params := &openapiverify.CreateVerificationCheckParams{}
	params.SetTo(to)
	params.SetCode(code)

	resp, err := client.VerifyV2.CreateVerificationCheck(servideId, params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Response: " + *resp.Status + " - " + *resp.To)
	}

}

func ListRateLimits(client *twilio.RestClient, serviceId string) {

	params := &openapiverify.ListRateLimitParams{}
	params.SetPageSize(10)
	resp, err := client.VerifyV2.ListRateLimit(serviceId, params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, rl := range resp.RateLimits {
			fmt.Println(*rl.UniqueName, *rl.Sid)
		}

	}
}

func CreateRateLimit(client *twilio.RestClient, serviceId string) {

	params := &openapiverify.CreateRateLimitParams{}
	params.SetUniqueName("LimitByCountryES")
	params.SetDescription("Limits in Spain")
	resp, err := client.VerifyV2.CreateRateLimit(serviceId, params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Created RateLimitID", resp.Sid)
	}
}

func CreateBucket(client *twilio.RestClient, serviceId string, rateLimitSid string) {

	params := &openapiverify.CreateBucketParams{}
	params.SetInterval(10 * 60)
	params.SetMax(1)
	resp, err := client.VerifyV2.CreateBucket(serviceId, rateLimitSid, params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(resp)
	}
}
