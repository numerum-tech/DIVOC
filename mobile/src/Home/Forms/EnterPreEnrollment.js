import {Button} from "react-bootstrap";
import React, {useState} from "react";
import {FORM_PRE_ENROLL_CODE, FORM_PRE_ENROLL_DETAILS, usePreEnrollment} from "./PreEnrollmentFlow";
import Form from "react-bootstrap/Form";
import "./EnterPreEnrollment.scss"
import {BaseFormCard} from "../../components/BaseFormCard";
import {useSelector} from "react-redux";
import {getMessageComponent, LANGUAGE_KEYS} from "../../lang/LocaleContext";

export const PHONE_NUMBER_MAX = 10

export function PreEnrollmentCode(props) {
    return (
        <BaseFormCard title={getMessageComponent(LANGUAGE_KEYS.PRE_ENROLLMENT_TITLE)}>
            <EnterPreEnrollmentContent/>
        </BaseFormCard>
    )
}

function EnterPreEnrollmentContent(props) {
    const {state, goNext} = usePreEnrollment()
    const [phoneNumber, setPhoneNumber] = useState(state.mobileNumber)
    const [enrollCode, setEnrollCode] = useState(state.enrollCode)
    const countryCode = useSelector(state => state.flagr.appConfig.countryCode);

    const handlePhoneNumberOnChange = (e) => {
        if (e.target.value.length <= PHONE_NUMBER_MAX) {
            setPhoneNumber(e.target.value)
        }
    }

    const handleEnrollCodeOnChange = (e) => {
        if (e.target.value.length <= 5) {
            setEnrollCode(e.target.value)
        }
    }
    return (
        <div className="enroll-code-container">
            <h4 className="title text-center">{getMessageComponent(LANGUAGE_KEYS.VERIFY_RECIPIENT_ENTER_MOBILE_AND_VERIFICATION_CODE)}</h4>
            <div className={"input-container"}>
                <div className="divOuter">
                    <div className="divInner">

                        <Form.Group>
                            <Form.Control type="text" placeholder={countryCode + "-XXXXXXXXX"} tabIndex="1"
                                          value={phoneNumber}
                                          onChange={handlePhoneNumberOnChange}/>
                            <Form.Control type="text" placeholder="XXXXX" tabIndex="1" value={enrollCode}
                                          onChange={handleEnrollCodeOnChange}/>
                            {/*<input id="otp" type="text" className="otp" tabIndex="2" maxLength="5"*/}
                            {/*       value={enrollCode}*/}
                            {/*       onChange={handleEnrollCodeOnChange}*/}
                            {/*       placeholder=""/>*/}
                        </Form.Group>

                    </div>
                </div>
            </div>
            <Button variant="outline-primary" className="action-btn" onClick={() => {
                goNext(FORM_PRE_ENROLL_CODE, FORM_PRE_ENROLL_DETAILS, {
                    mobileNumber: phoneNumber,
                    enrollCode: enrollCode
                })
            }}>{getMessageComponent(LANGUAGE_KEYS.VERIFY_RECIPIENT_CONFIRM_BUTTON)}</Button>
        </div>
    );
}
