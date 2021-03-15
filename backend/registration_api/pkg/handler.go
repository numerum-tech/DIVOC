package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/divoc/kernel_library/model"
	kernelService "github.com/divoc/kernel_library/services"
	"github.com/divoc/registration-api/config"
	"github.com/divoc/registration-api/pkg/enrollment"
	models2 "github.com/divoc/registration-api/pkg/models"
	"github.com/divoc/registration-api/pkg/services"
	"github.com/divoc/registration-api/pkg/utils"
	models3 "github.com/divoc/registration-api/swagger_gen/models"
	"github.com/divoc/registration-api/swagger_gen/restapi/operations"
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const FacilityEntity = "Facility"
const EnrollmentEntity = "Enrollment"
const LastInitializedKey = "LAST_FACILITY_SLOTS_INITIALIZED"
const YYYYMMDD = "2006-01-02"
const AttemptsKey = "attempts"
const OtpKey = "otp"

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
	api.DeleteAppointmentHandler = operations.DeleteAppointmentHandlerFunc(deleteAppointment)
	api.DeleteRecipientHandler = operations.DeleteRecipientHandlerFunc(deleteRecipient)
	api.GetPingHandler = operations.GetPingHandlerFunc(pingHandler)
}

func pingHandler(params operations.GetPingParams) middleware.Responder {
	return operations.NewGetPingOK()
}

func getRecipients(params operations.GetRecipientsParams, principal *models3.JWTClaimBody) middleware.Responder {
	filter := map[string]interface{}{}
	filter["phone"] = map[string]interface{}{
		"eq": principal.Phone,
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

func enrollRecipient(params operations.EnrollRecipientParams, principal *models3.JWTClaimBody) middleware.Responder {
	params.Body.Phone = principal.Phone
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

	if _, err := services.SetHMSet(phone, map[string]interface{}{OtpKey: otp, AttemptsKey: 0}); err == nil {
		if _, err := services.SetTTLForHash(phone, time.Minute*time.Duration(config.Config.Auth.TTLForOtp)); err == nil {
			if config.Config.MockOtp {
				return operations.NewGenerateOTPOK()
			}
			// Send SMS
			if _, err := utils.SendOTP("+91", phone, otp); err == nil {
				return operations.NewGenerateOTPOK()
			} else {
				log.Errorf("Error while sending OTP %+v", err)
				return operations.NewGenerateOTPInternalServerError()
			}
		} else {
			log.Errorf("Error occurred while trying to set ttl for hash %+v", err)
			return operations.NewGenerateOTPInternalServerError()
		}
	} else {
		log.Errorf("Error occurred while trying to cache the otp details %+v", err)
		return operations.NewGenerateOTPInternalServerError()
	}
}

func verifyOTP(params operations.VerifyOTPParams) middleware.Responder {
	phone := params.Body.Phone
	receivedOTP := params.Body.Otp
	if receivedOTP == "" {
		return operations.NewVerifyOTPBadRequest()
	}
	otpDetails, err := services.GetHashValues(phone)
	if err != nil {
		log.Errorf("No OTP in the store, might have expired. %+v", err)
		return operations.NewVerifyOTPUnauthorized()
	}

	if attemptsTried, err := services.IncrHashField(phone, AttemptsKey); err == nil {
		if attemptsTried > config.Config.Auth.MAXOtpVerifyAttempts {
			if err = services.DeleteValue(phone); err != nil {
				log.Errorf("Error in clearing the OTP in redis %+v", err)
				return model.NewGenericServerError()
			}
			return operations.NewVerifyOTPTooManyRequests()
		}
	}

	if otpDetails[OtpKey] != receivedOTP {
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
	currentDate := time.Now()
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
						services.ClearOldSlots(facilityCode, currentDate.Unix())
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
											for i := 1; i < config.Config.AppointmentScheduler.ScheduleDays; i++ {
												slotDate := currentDate.AddDate(0, 0, i)
												programSchedulesForDay, isFacilityAvailableForSlot := programSchedule[slotDate.Weekday()]
												for _, programSchedule := range programSchedulesForDay {
													if isFacilityAvailableForSlot {
														startTime := programSchedule["startTime"]
														endTime := programSchedule["endTime"]
														maxAppointments := programSchedule["maxAppointments"]
														schedule := models2.FacilitySchedule{
															FacilityCode: facilityCode,
															ProgramId:    programId,
															Date:         slotDate,
															StartTime:    startTime,
															EndTime:      endTime,
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

func getFacilitySlots(params operations.GetSlotsForFacilitiesParams, principal *models3.JWTClaimBody) middleware.Responder {
	if params.FacilityID == nil {
		return operations.NewGenerateOTPBadRequest()
	}
	offset := (*params.PageNumber) * (*params.PageSize)
	now := fmt.Sprintf("%d", time.Now().Unix())
	slotKeys, err := services.GetValuesByScoreFromSet(*params.FacilityID, now, "inf", *params.PageSize, offset)
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

func bookSlot(params operations.BookSlotOfFacilityParams, principal *models3.JWTClaimBody) middleware.Responder {
	if params.Body.EnrollmentCode == nil || params.Body.FacilitySlotID == nil {
		return operations.NewBookSlotOfFacilityBadRequest()
	}

	enrollmentInfo := getEnrollmentInfoIfValid(*params.Body.EnrollmentCode, principal.Phone)
	if enrollmentInfo != nil {
		if !checkIfAlreadyAppointed(enrollmentInfo) {
			err := services.BookAppointmentSlot(*params.Body.FacilitySlotID)
			if err != nil {
				return operations.NewBookSlotOfFacilityBadRequest()
			} else {
				isMarked := services.MarkEnrollmentAsBooked(*params.Body.EnrollmentCode, *params.Body.FacilitySlotID)
				if isMarked {
					facilitySchedule := models2.ToFacilitySchedule(*params.Body.FacilitySlotID)
					services.PublishAppointmentAcknowledgement(models2.AppointmentAck{
						Dose:            *params.Body.Dose,
						ProgramId:       *params.Body.ProgramID,
						EnrollmentCode:  *params.Body.EnrollmentCode,
						SlotID:          *params.Body.FacilitySlotID,
						FacilityCode:    facilitySchedule.FacilityCode,
						AppointmentDate: facilitySchedule.DateString(),
						AppointmentTime: facilitySchedule.StartTime + "-" + facilitySchedule.EndTime,
						CreatedAt:       time.Now(),
						Status:          models2.AllottedStatus,
					})

					return operations.NewGetSlotsForFacilitiesOK()
				}
			}
		} else {
			log.Errorf("Already booked %s, %s", *params.Body.EnrollmentCode, principal.Phone)
		}
	} else {
		log.Errorf("Invalid booking request %s, %s", *params.Body.EnrollmentCode, principal.Phone)
	}
	return operations.NewGetSlotsForFacilitiesBadRequest()
}

func checkIfAlreadyAppointed(enrollmentInfo map[string]string) bool {
	if _, ok := enrollmentInfo["slotId"]; ok {
		return true
	}
	return false
}

func getEnrollmentInfoIfValid(enrollmentCode string, phone string) map[string]string {
	values, err := services.GetHashValues(enrollmentCode)
	if err == nil {
		if val, ok := values["phone"]; ok && val == phone {
			return values
		}
	}
	return nil
}

func deleteAppointment(params operations.DeleteAppointmentParams, principal *models3.JWTClaimBody) middleware.Responder {
	if params.Body.EnrollmentCode == nil {
		return operations.NewDeleteAppointmentBadRequest()
	}

	deleteError := deleteAppointmentInEnrollment(*params.Body.EnrollmentCode, principal.Phone, *params.Body.Dose, *params.Body.ProgramID)
	if deleteError == nil {
		return operations.NewDeleteRecipientOK()
	} else {
		errorMessage := deleteError.Error()
		response := operations.NewDeleteAppointmentBadRequest()
		response.Payload = &operations.DeleteAppointmentBadRequestBody{
			Message: errorMessage,
		}
		log.Info(errorMessage)
		return response
	}
}

func deleteAppointmentInEnrollment(enrollmentCode string, phone string, dose string, programId string) error {
	enrollmentInfo := getEnrollmentInfoIfValid(enrollmentCode, phone)
	if enrollmentInfo != nil {
		if checkIfAlreadyAppointed(enrollmentInfo) {
			if msg := checkIfCancellationAllowed(enrollmentInfo); msg == "" {
				lastBookedSlotId := enrollmentInfo["slotId"]
				err := services.CancelBookedAppointment(lastBookedSlotId)
				if err != nil {
					return errors.New("Failed to cancel appointment")
				} else {
					isMarked := services.RevokeEnrollmentBookedStatus(enrollmentCode)
					if isMarked {
						services.PublishAppointmentAcknowledgement(models2.AppointmentAck{
							EnrollmentCode:  enrollmentCode,
							Dose:            dose,
							ProgramId:       programId,
							SlotID:          "",
							FacilityCode:    "",
							AppointmentDate: "0001-01-01",
							AppointmentTime: "",
							CreatedAt:       time.Now(),
							Status:          models2.CancelledStatus,
						})
						return nil
					}
				}
			} else {
				log.Errorf("Cancellation of appointment not allowed %v", msg)
				return errors.New(msg)
			}
		} else {
			return errors.New("Enrollment not booked " + enrollmentCode + "," + phone)
		}
	} else {
		return errors.New("Invalid booking request  " + enrollmentCode + "," + phone)
	}
	return errors.New("Failed to cancel appointment")
}

func deleteRecipient(params operations.DeleteRecipientParams, principal *models3.JWTClaimBody) middleware.Responder {
	enrollmentCode := *params.Body.EnrollmentCode
	enrollmentInfo := getEnrollmentInfoIfValid(enrollmentCode, principal.Phone)
	if enrollmentInfo != nil && !checkIfAlreadyAppointed(enrollmentInfo) {
		if osid, ok := enrollmentInfo["osid"]; ok {
			err := enrollment.DeleteRecipient(osid)
			if err != nil {
				log.Error(err)
				return operations.NewDeleteRecipientBadRequest()
			} else {
				err := services.DeleteValue(enrollmentCode)
				log.Error(err)
			}
			return operations.NewDeleteRecipientOK()
		}
	} else {
		errorMessage := "Deleting a recipient is not allowed if appointment is scheduled."
		response := operations.NewDeleteAppointmentBadRequest()
		response.Payload = &operations.DeleteAppointmentBadRequestBody{
			Message: errorMessage,
		}
		log.Info(errorMessage)
		return response
	}
	return operations.NewDeleteRecipientBadRequest()
}

func checkIfCancellationAllowed(enrollmentInfo map[string]string) string {
	lastBookedSlotId := enrollmentInfo["slotId"]
	facilitySchedule := models2.ToFacilitySchedule(lastBookedSlotId)
	remainingHoursForSchedule := facilitySchedule.Date.Sub(time.Now()).Hours()
	if remainingHoursForSchedule <= 0 {
		return fmt.Sprintf("Cancellation is not allowed")
	}
	if remainingHoursForSchedule <= float64(config.Config.MinCancellationHours) {
		return fmt.Sprintf("Cancellation before %d hours is not allowed", config.Config.MinCancellationHours)
	}
	updatedCount, _ := strconv.Atoi(enrollmentInfo["updatedCount"])
	if updatedCount >= config.Config.MaxAppointmentUpdatesAllowed {
		return fmt.Sprintf("You have reached the maximum number of times to update appointment")
	}
	return ""
}
