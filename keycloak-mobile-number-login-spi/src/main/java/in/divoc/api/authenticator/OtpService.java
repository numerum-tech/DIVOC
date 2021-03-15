package in.divoc.api.authenticator;

import org.jboss.logging.Logger;

import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.http.RequestBody;

import java.util.Random;

class OtpService {
    private static final Logger logger = Logger.getLogger(OtpService.class);

    String sendOtp(String mobileNumber) {
        Random rand = new Random();
        String otp = String.format("%04d", rand.nextInt(10000));
        logger.infov("OTP {0} is sent for mobile number {1}", otp, mobileNumber);
        //String otp = "1234";
        try {
            togoSendOtp(mobileNumber, otp);
            return otp;
        } catch (Exception e) {
            //TODO: handle exception
            logger.infov("OTP {0} fail sending to mobile number {1}  {2}", otp, mobileNumber, e.getMessage());
            return null;
        }
    }

    private void togoSendOtp(String mobileNumber, String text) throws Exception {

        // form parameters
        RequestBody formBody = new FormBody.Builder()
                .add("id", "5747569069")
                .add("msisdn", mobileNumber)
                .add("message", text)
                .build();
    

        Request request = new Request.Builder()
                .url("https://sms.smarthub.gouv.tg/api/peers/send_sms")
                .addHeader("User-Agent", "OkHttp Bot")
                .post(formBody)
                .build();

        HttpClient httpClient = HttpClient.newHttpClient();

        try (Response response = httpClient.newCall(request).execute()) {

            if (!response.isSuccessful()) throw new IOException("Unexpected code " + response);

            // Get response body
            //System.out.println(response.body().string());
        }

    }
}
