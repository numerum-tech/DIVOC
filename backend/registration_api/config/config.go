package config

import (
	"github.com/jinzhu/configor"
)

var Config = struct {
	Auth struct {
		PublicKey            string `yaml:"publickey" env:"AUTH_PUBLIC_KEY"`
		PrivateKey           string `yaml:"privatekey" env:"AUTH_PRIVATE_KEY"`
		TTLForOtp            int    `yaml:"ttlforotpinminutes"`
		MAXOtpVerifyAttempts int    `yaml:"maxotpverifyattempts"`
		OTPLength			 int    `yaml:"otp_length" env:"OTP_LENGTH" default:"4"`
	}

	Kafka struct {
		BootstrapServers    string `env:"KAFKA_BOOTSTRAP_SERVERS" yaml:"bootstrapservers"`
		NotifyTopic         string `default:"notify" yaml:"notifyTopic"`
		EnrollmentTopic     string `default:"enrollment" yaml:"enrollmenttopic"`
		AppointmentAckTopic string `default:"appointment_ack" yaml:"appointmentacktopic"`
		RecipientAppointmentTopic string `default:"recipientappointment" yaml:"recipientappointmenttopic"`
	}

	EnrollmentCreation struct {
		MaxRetryCount                  int `default:"10" yaml:"maxretrycount"`
		LengthOfSuffixedEnrollmentCode int `default:"10" yaml:"lengthofsuffixedenrollmentcode"`
	}
	Redis struct {
		Url             string  `env:"REDIS_URL"`
		CacheTTL		int     `default:"60" env:"CACHE_TTL"`
	}
	AppointmentScheduler struct {
		ChannelSize    int `default:"100"`
		ChannelWorkers int `default:"10"`
		ScheduleDays   int `default:"30"`
	}
	MockOtp bool `default:"true" env:"MOCK_OTP"`
}{}

func Initialize() {
	err := configor.Load(&Config, "./config/application-default.yml") //"config/application.yml"

	if err != nil {
		panic("Unable to read configurations")
	}
}
