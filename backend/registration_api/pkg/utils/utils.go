package utils

import (
	"github.com/divoc/registration-api/config"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func GenerateEnrollmentCode(phoneNumber string) string {
	log.Info("Generating the code for the length : ",
		config.Config.EnrollmentCreation.LengthOfSuffixedEnrollmentCode, config.Config.EnrollmentCreation.MaxRetryCount)
	digits := 0

	n:= config.Config.EnrollmentCreation.LengthOfSuffixedEnrollmentCode
	for ;n>=1;n-- {
		digits  = digits * 10 + 9
	}
	return phoneNumber + "-" + strconv.Itoa(rand.Intn(digits))
}

func GenerateOTP() string {
	if config.Config.MockOtp {
		return "1234"
	} else {
		n := config.Config.Auth.OTPLength
		otp := int(math.Pow10(n-1)) + rand.Intn(int(math.Pow10(n)-math.Pow10(n-1)))
		return strconv.Itoa(otp)
	}
}

//todo: move to notification service with priority as transactional
func SendOTP(prefix string, phone string, otp string) (*sns.PublishOutput, error) {
	sess := session.Must(session.NewSession())
	log.Info("session created")
	svc := sns.New(sess)
	log.Info("service created")
	msgType := "Transactional"
	dataType := "String"
	svc.SetSMSAttributesRequest(&sns.SetSMSAttributesInput{
		Attributes: map[string]*string{"DefaultSMSType": &msgType,},
	})
	params := &sns.PublishInput{
		Message: aws.String("OTP for registration " + otp),
		PhoneNumber: aws.String(prefix + phone),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"AWS.SNS.SMS.SMSType": &sns.MessageAttributeValue{
				DataType:    &dataType,
				StringValue: &msgType,
			},
		},
	}
	resp, err := svc.Publish(params)
	log.Infof("Message sent %s %+v", phone, resp)
	return resp, err
}