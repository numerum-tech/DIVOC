package consumers

import (
	"encoding/json"
	kernelService "github.com/divoc/kernel_library/services"
	"github.com/divoc/registration-api/config"
	"github.com/divoc/registration-api/pkg/services"
	"github.com/divoc/registration-api/swagger_gen/models"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"time"
)

func StartRecipientsAppointmentBookingConsumer() {
	servers := config.Config.Kafka.BootstrapServers
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  servers,
		"group.id":           "recipient-appointment",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		log.Errorf("Connection failed for the kafka",  err)
		return
	}

	go func() {
		err := consumer.SubscribeTopics([]string{config.Config.Kafka.RecipientAppointmentTopic}, nil)
		if err != nil {
			panic(err)
		}
		for {
			msg, err := consumer.ReadMessage(-1)
			if err == nil {
				log.Info("Got the message to book an appointment")
				var recipientAppointmentMessage RecipientAppointmentMessage
				err = json.Unmarshal(msg.Value, &recipientAppointmentMessage)
				if err == nil {
					filter := map[string]interface{}{}
					filter["code"] = map[string]interface{}{
						"eq": recipientAppointmentMessage.EnrollmentCode,
					}
					if responseFromRegistry, err := kernelService.QueryRegistry("Enrollment", filter, 100, 0); err==nil {
						appointmentTData := make(map[string]interface{})
						enrollment := responseFromRegistry["Enrollment"].([]interface{})[0].(map[string]interface{})

						appointmentTData["appointmentSlot"] = recipientAppointmentMessage.AppointmentTime
						appointmentTData["osid"] = enrollment["osid"]
						appointmentTData["appointmentDate"] = recipientAppointmentMessage.AppointmentDate
						appointmentTData["enrollmentScopeId"] = recipientAppointmentMessage.FacilityCode

						log.Infof("Message on %s: %v \n", msg.TopicPartition, appointmentTData)

						_, err := kernelService.UpdateRegistry("Enrollment", appointmentTData)
						if err == nil {
							err := services.NotifyAppointmentBooked(CreateEnrollmentFromInterface(enrollment, recipientAppointmentMessage.AppointmentDate,
								recipientAppointmentMessage.AppointmentTime))
							if err != nil {
								log.Error("Unable to send notification to the enrolled user",  err)
							}
						} else {
							// Push to error topic
							log.Error("Booking appointment is failed ", err)
						}
						_, _ = consumer.CommitMessage(msg)
					} else {
						log.Errorf("Unable to fetch the osid for Enrollment", err)
					}
				} else {
					log.Info("Unable to serialize the request body", err)
				}

			} else {
				// The client will automatically try to recover from all errors.
				log.Infof("Consumer error: %v \n", err)
			}
		}
	}()
}

func CreateEnrollmentFromInterface(enrollmentMap map[string]interface{}, appointmentDate string, appointmentTime string) models.Enrollment {

	appointmentDateFormat, err := time.Parse("2006-01-02", appointmentDate)
	if err != nil {
		log.Errorf("Invalid date format (%v) (%v)", appointmentDate, err)
		return models.Enrollment{}
	} else {
		return models.Enrollment{
			Name:            enrollmentMap["name"].(string),
			Phone:           enrollmentMap["phone"].(string),
			Email:           enrollmentMap["email"].(string),
			AppointmentDate: strfmt.Date(appointmentDateFormat),
			AppointmentSlot: appointmentTime,
		}
	}
}

type RecipientAppointmentMessage struct {
	EnrollmentCode  string
	FacilityCode    string
	AppointmentDate string
	AppointmentTime string
}
