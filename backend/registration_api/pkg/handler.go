package pkg

import (
	"encoding/json"
	"github.com/divoc/kernel_library/model"
	kernelService "github.com/divoc/kernel_library/services"
	"github.com/divoc/registration-api/config"
	"github.com/divoc/registration-api/models"
	models2 "github.com/divoc/registration-api/pkg/models"
	"github.com/divoc/registration-api/pkg/services"
	"github.com/divoc/registration-api/pkg/utils"
	"github.com/divoc/registration-api/swagger_gen/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const FacilityEntity = "Facility"
const EnrollmentEntity = "Enrollment"
const LastInitializedKey = "LAST_FACILITY_SLOTS_INITIALIZED"
const YYYYMMDD = "2006-01-02"
const SlotsToReturn = 100

var DaysMap = map[string]time.Weekday{
	"Su": time.Sunday,
	"Mo": time.Monday,
	"Tu": time.Tuesday,
	"We": time.Wednesday,
	"Th": time.Thursday,
	"Fr": time.Friday,
	"Sa": time.Saturday,
}

func SetupHandlers(api *operations.RegistrationAPIAPI) {
	api.EnrollRecipientHandler = operations.EnrollRecipientHandlerFunc(enrollRecipient)
	api.GenerateOTPHandler = operations.GenerateOTPHandlerFunc(generateOTP)
	api.VerifyOTPHandler = operations.VerifyOTPHandlerFunc(verifyOTP)
	api.GetRecipientsHandler = operations.GetRecipientsHandlerFunc(getRecipients)
	api.InitializeFacilitySlotsHandler = operations.InitializeFacilitySlotsHandlerFunc(initializeFacilitySlots)
	api.GetSlotsForFacilitiesHandler = operations.GetSlotsForFacilitiesHandlerFunc(getFacilitySlots)
	api.BookSlotOfFacilityHandler = operations.BookSlotOfFacilityHandlerFunc(bookSlot)
}

func getRecipients(params operations.GetRecipientsParams) middleware.Responder {
	recipientToken := params.HTTPRequest.Header.Get("recipientToken")
	if recipientToken == "" {
		log.Error("Recipient Token is empty")
		return operations.NewGetRecipientsUnauthorized()
	}
	phone, err := services.VerifyRecipientToken(recipientToken)
	if err != nil {
		log.Error("Error occurred while verifying the token ", err)
		return operations.NewGetRecipientsUnauthorized()
	}
	filter := map[string]interface{}{}
	filter["phone"] = map[string]interface{}{
		"eq": phone,
	}
	responseFromRegistry, err := kernelService.QueryRegistry(EnrollmentEntity, filter, 100, 0)
	if err != nil {
		log.Error("Error occurred while querying Enrollment registry ", err)
		return operations.NewGetRecipientsInternalServerError()
	}
	if enrollmentArr, err := json.Marshal(responseFromRegistry["Enrollment"]); err == nil {
		var enrollments []map[string]interface{}
		err := json.Unmarshal(enrollmentArr, &enrollments)
		if err != nil {
			log.Errorf("Error occurred while trying to unmarshal the array of enrollments (%v)", err)
			return model.NewGenericServerError()
		} else {
			services.EnrichFacilityDetails(enrollments)
			return model.NewGenericJSONResponse(enrollments)
		}
	} else {
		log.Errorf("Error occurred while trying to marshal the array of enrollments (%v)", err)
		return model.NewGenericServerError()
	}
}

func enrollRecipient(params operations.EnrollRecipientParams) middleware.Responder {
	recipientToken := params.HTTPRequest.Header.Get("recipientToken")
	if recipientToken == "" {
		log.Error("Recipient Token is empty")
		return operations.NewEnrollRecipientUnauthorized()
	}
	phone, err := services.VerifyRecipientToken(recipientToken)
	if err != nil {
		log.Error("Error occurred while verifying the token ", err)
		return operations.NewEnrollRecipientUnauthorized()
	}
	params.Body.Phone = phone
	if recipientData, err := json.Marshal(params.Body); err == nil {
		log.Info("Received Recipient data to enroll", string(recipientData), params.Body)
		services.PublishEnrollmentMessage(recipientData)
	}
	return operations.NewEnrollRecipientOK()
}

func generateOTP(params operations.GenerateOTPParams) middleware.Responder {
	phone := params.Body.Phone
	if phone == "" {
		return operations.NewGenerateOTPBadRequest()
	}
	otp := utils.GenerateOTP()
	cacheOtp, err := json.Marshal(models.CacheOTP{Otp: otp, VerifyAttemptCount: 0})
	err = services.SetValue(phone, string(cacheOtp), time.Duration(config.Config.Auth.TTLForOtp))
	if config.Config.MockOtp {
		return operations.NewGenerateOTPOK()
	}
	if err == nil {
		// Send SMS
		if _, err := utils.SendOTP("+91", phone, otp); err == nil {
			return operations.NewGenerateOTPOK()
		} else {
			log.Errorf("Error while sending OTP %+v", err)
			return operations.NewGenerateOTPInternalServerError()
		}
	} else {
		log.Errorf("Error while setting otp in redis %+v", err)
		return operations.NewGenerateOTPInternalServerError()
	}
}

func verifyOTP(params operations.VerifyOTPParams) middleware.Responder {
	phone := params.Body.Phone
	receivedOTP := params.Body.Otp
	if receivedOTP == "" {
		return operations.NewVerifyOTPBadRequest()
	}
	value, err := services.GetValue(phone)
	if err != nil {
		return model.NewGenericServerError()
	}
	if value == "" {
		return operations.NewVerifyOTPUnauthorized()
	}

	cacheOTP := models.CacheOTP{}
	if err := json.Unmarshal([]byte(value), &cacheOTP); err != nil {
		log.Errorf("Error in marshalling json %+v", err)
		return model.NewGenericServerError()
	}
	if cacheOTP.VerifyAttemptCount > config.Config.Auth.MAXOtpVerifyAttempts {
		return operations.NewVerifyOTPTooManyRequests()
	}
	if cacheOTP.Otp != receivedOTP {
		cacheOTP.VerifyAttemptCount += 1
		if cacheOtp, err := json.Marshal(cacheOTP); err != nil {
			log.Errorf("Error in setting verify count %+v", err)
		} else {
			err = services.SetValue(phone, string(cacheOtp), time.Duration(config.Config.Auth.TTLForOtp))
		}
		return operations.NewVerifyOTPUnauthorized()
	}

	if err = services.DeleteValue(phone); err != nil {
		log.Errorf("Error in clearing the OTP  after signin %+v", err)
		return model.NewGenericServerError()
	} else {
		token, err := services.CreateRecipientToken(phone)
		if err != nil {
			log.Errorf("Unable to create the jwt token %+v", err)
			return model.NewGenericServerError()
		}
		response := operations.VerifyOTPOKBody{
			Token: token,
		}
		return operations.NewVerifyOTPOK().WithPayload(&response)
	}
}

func canInitializeSlots() bool {
	lastInitializedDate, err := services.GetValue(LastInitializedKey)
	if err != nil {
		return true
	} else {
		initializedDate, _ := time.Parse(YYYYMMDD, lastInitializedDate)
		currentDate := time.Now()
		if initializedDate.YearDay() == currentDate.YearDay() && initializedDate.Year() == currentDate.Year() {
			return false
		}
		return true
	}
}

func initializeFacilitySlots(params operations.InitializeFacilitySlotsParams) middleware.Responder {

	if canInitializeSlots() {
		log.Infof("Initializing facility slots")
		filters := map[string]interface{}{}
		limit := 1000
		offset := -1000
		for {
			offset += limit
			facilitiesResponse, err := kernelService.QueryRegistry(FacilityEntity, filters, limit, offset)
			facilities, ok := facilitiesResponse[FacilityEntity].([]interface{})
			if err != nil || !ok {
				if err != nil {
					log.Error("Fetching facilities failed", err)
				}
				return operations.NewGenerateOTPBadRequest()
			} else if len(facilities) == 0 {
				_ = services.SetValueWithoutExpiry(LastInitializedKey, time.Now().Format(YYYYMMDD))
				return operations.NewInitializeFacilitySlotsOK()
			} else {
				for _, facilityObj := range facilities {
					facility, ok := facilityObj.(map[string]interface{})
					if ok {
						facilityCode := facility["facilityCode"].(string)
						facilityOSID := facility["osid"].(string)
						log.Infof("Initializing facility %s slots", facilityCode)

						facilityProgramArr, ok := facility["programs"].([]interface{})
						facilityProgramWiseSchedule := services.GetFacilityAppointmentSchedule(facilityOSID)
						if ok && len(facilityProgramArr) > 0 {
							for _, facilityProgramObj := range facilityProgramArr {
								facilityProgram, ok := facilityProgramObj.(map[string]interface{})
								if ok {
									programId, ok := facilityProgram["programId"].(string)
									programStatus, ok := facilityProgram["status"].(string)
									if ok && programStatus == "Active" {
										programSchedule, ok := facilityProgramWiseSchedule[programId]
										if ok {
											currentDate := time.Now()
											for i := 0; i < config.Config.AppointmentScheduler.ScheduleDays; i++ {
												slotDate := currentDate.AddDate(0, 0, i)
												programSchedulesForDay, isFacilityAvailableForSlot := programSchedule[slotDate.Weekday()]
												for _, programSchedule := range programSchedulesForDay {
													if isFacilityAvailableForSlot {
														startTime := programSchedule["startTime"]
														endTime := programSchedule["endTime"]
														maxAppointments := programSchedule["maxAppointments"]
														schedule := services.FacilitySchedule{
															FacilityCode: facilityCode,
															ProgramId:    programId,
															Date:         slotDate,
															Time:         startTime + "_" + endTime,
															Slots:        maxAppointments,
														}
														log.Infof("Initializing facility slot %v", schedule)
														services.AddFacilityScheduleToChannel(schedule)
														log.Infof("Initialized facility slot %v", schedule)
													}
												}
											}
										}
									}
								}
							}
						}

					}
				}
			}
		}
	}
	return operations.NewInitializeFacilitySlotsUnauthorized()
}

func getFacilitySlots(paras operations.GetSlotsForFacilitiesParams) middleware.Responder {
	if paras.FacilityID == nil {
		return operations.NewGenerateOTPBadRequest()
	}
	startPosition := int64(*paras.PageNumber) * SlotsToReturn
	slotKeys, err := services.GetValuesFromSet(*paras.FacilityID, startPosition, startPosition+SlotsToReturn-1)
	if err == nil && len(slotKeys) > 0 {
		slotsAvailable, err := services.GetValues(slotKeys...)
		if err == nil {
			return &operations.GetSlotsForFacilitiesOK{
				Payload: map[string]interface{}{
					"keys":  slotKeys,
					"slots": slotsAvailable,
				},
			}
		}
	}
	return operations.NewGetSlotsForFacilitiesBadRequest()
}

func bookSlot(params operations.BookSlotOfFacilityParams) middleware.Responder {
	recipientToken := params.HTTPRequest.Header.Get("recipientToken")
	if params.Body.EnrollmentCode == nil || params.Body.FacilitySlotID == nil {
		return operations.NewBookSlotOfFacilityBadRequest()
	}
	if recipientToken == "" {
		log.Error("Recipient Token is empty")
		return operations.NewGetRecipientsUnauthorized()
	}
	phone, err := services.VerifyRecipientToken(recipientToken)
	if err != nil {
		log.Error("Invalid Token")
		return operations.NewGetRecipientsUnauthorized()
	}
	if isValidEnrollmentCode(*params.Body.EnrollmentCode, phone) {
		if !checkIfAlreadyAppointed(*params.Body.EnrollmentCode) {
			err := services.BookAppointmentSlot(*params.Body.FacilitySlotID)
			if err != nil {
				return operations.NewBookSlotOfFacilityBadRequest()
			} else {
				isMarked := services.MarkEnrollmentCodeAsBooked(*params.Body.EnrollmentCode, *params.Body.FacilitySlotID)
				if isMarked {
					facilityDetails := strings.Split(*params.Body.FacilitySlotID, "_")
					services.PublishAppointmentAcknowledgement(models2.AppointmentAck{
						EnrollmentCode:  *params.Body.EnrollmentCode,
						SlotID:          *params.Body.FacilitySlotID,
						FacilityCode:    facilityDetails[0],
						AppointmentDate: facilityDetails[2],
						AppointmentTime: facilityDetails[3] + "-" + facilityDetails[4],
						CreatedAt:       time.Now(),
					})

					return operations.NewGetSlotsForFacilitiesOK()
				}
			}
		} else {
			log.Errorf("Already booked %s, %s", *params.Body.EnrollmentCode, phone)
		}
	} else {
		log.Errorf("Invalid booking request %s, %s", *params.Body.EnrollmentCode, phone)
	}
	return operations.NewGetSlotsForFacilitiesBadRequest()
}

func checkIfAlreadyAppointed(enrollmentCode string) bool {
	exists, err := services.HashFieldExists(enrollmentCode, "slotId")
	if err != nil {
		return false
	} else {
		return exists
	}
}

func isValidEnrollmentCode(enrollmentCode string, phone string) bool {
	//TODO: check in cache, store the registered enrolled users in cache as map
	filter := map[string]interface{}{
		"code": map[string]interface{}{
			"eq": enrollmentCode,
		},
		"phone": map[string]interface{}{
			"eq": phone,
		},
	}
	enrollmentsArr, err := kernelService.QueryRegistry(EnrollmentEntity, filter, 100, 0)
	if err == nil {
		enrollments := enrollmentsArr[EnrollmentEntity].([]interface{})
		if len(enrollments) > 0 {
			return true
		}
	}
	return false
}
