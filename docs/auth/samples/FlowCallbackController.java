import org.springframework.web.bind.annotation.*;
import org.springframework.http.*;
import org.springframework.web.client.RestTemplate;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.util.HexFormat;

/**
 * Example Spring Boot Controller handling the Flow Director callback and
 * Authentication.
 * 
 * This sample shows:
 * 1. API Key Generation (Team initialization)
 * 2. Receiving a request from the Director with an 'X-Callback-Secret'.
 * 3. Signing the response body using HMAC-SHA256 with the secret.
 * 4. Sending a response back to the 'callbackURL' including the 'X-Signature'.
 */
@RestController
@RequestMapping("/api/flow")
public class FlowCallbackController {
    private final RestTemplate restTemplate = new RestTemplate();
    private final ObjectMapper objectMapper = new ObjectMapper();

    /**
     * Utility to generate the API Key for the initial token exchange.
     * The students should use this to generate the 'apiKey' they send to
     * /auth/token.
     */
    public String generateInitialApiKey(String clientID, String secret) throws Exception {
        return calculateHmac(clientID, secret);
    }

    /**
     * The Director calls this endpoint when action is required.
     * 
     * @param callbackSecret The secret provided by the Director for this specific
     *                       request.
     * @param body           The request body containing session info and the
     *                       'callbackUrl'.
     */
    @PostMapping("/callback")
    public ResponseEntity<String> handleDirectorCallback(
            @RequestHeader(value = "X-Callback-Secret", required = false) String callbackSecret,
            @RequestBody String body) {

        try {
            System.out.println("Received request from Director: " + body);
            JsonNode jsonNode = objectMapper.readTree(body);

            // The Director provides a unique callback URL for each request
            String callbackUrl = jsonNode.get("callback").asText();

            // --- Business Logic Start ---
            // Simulate processing...
            String moveResponse = "{\"move\": \"attack\", \"target\": \"player2\"}";
            // --- Business Logic End ---

            // To respond, we MUST sign the body with the secret and include it in
            // 'X-Signature'
            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);

            if (callbackSecret != null) {
                // IMPORTANT: Signature is HMAC-SHA256(Secret, Body)
                String signature = calculateHmac(moveResponse, callbackSecret);
                headers.set("X-Signature", signature);
                System.out.println("Generated X-Signature for callback: " + signature);
            }

            HttpEntity<String> request = new HttpEntity<>(moveResponse, headers);

            // Sending the response back to the Director
            ResponseEntity<String> response = restTemplate.postForEntity(callbackUrl, request, String.class);

            if (response.getStatusCode().is2xxSuccessful()) {
                System.out.println("Callback successfully delivered to Director.");
            }

            return ResponseEntity.accepted().body("Action handled and callback sent.");

        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.internalServerError().body("Error processing callback: " + e.getMessage());
        }
    }

    /**
     * Helper to calculate HMAC-SHA256 signature.
     */
    private String calculateHmac(String data, String key) throws Exception {
        Mac sha256_HMAC = Mac.getInstance("HmacSHA256");
        SecretKeySpec secret_key = new SecretKeySpec(key.getBytes("UTF-8"), "HmacSHA256");
        sha256_HMAC.init(secret_key);
        return HexFormat.of().formatHex(sha256_HMAC.doFinal(data.getBytes("UTF-8")));
    }
}
