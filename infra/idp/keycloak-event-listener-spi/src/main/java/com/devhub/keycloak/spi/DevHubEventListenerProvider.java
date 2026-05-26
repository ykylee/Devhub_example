package com.devhub.keycloak.spi;

import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.admin.AdminEvent;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.logging.Level;
import java.util.logging.Logger;

public class DevHubEventListenerProvider implements EventListenerProvider {

    private static final Logger logger = Logger.getLogger(DevHubEventListenerProvider.class.getName());
    private static final HttpClient httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(3))
            .build();

    private final String webhookUrl;
    private final String webhookSecret;

    public DevHubEventListenerProvider(String webhookUrl, String webhookSecret) {
        this.webhookUrl = webhookUrl;
        this.webhookSecret = webhookSecret;
    }

    @Override
    public void onEvent(Event event) {
        if (event == null) return;
        // Skip common noisy user events if necessary (e.g. REFRESH_TOKEN) to reduce webhook load
        String type = event.getType() != null ? event.getType().toString() : "UNKNOWN";
        if ("REFRESH_TOKEN".equals(type) || "CODE_TO_TOKEN".equals(type) || "INTROSPECT_TOKEN".equals(type)) {
            return;
        }

        String jsonPayload = String.format(
            "{" +
            "\"id\":\"%s\"," +
            "\"time\":%d," +
            "\"type\":\"%s\"," +
            "\"realmId\":\"%s\"," +
            "\"clientId\":\"%s\"," +
            "\"userId\":\"%s\"," +
            "\"ipAddress\":\"%s\"," +
            "\"error\":\"%s\"," +
            "\"details\":{}" +
            "}",
            escapeJson(event.getId()),
            event.getTime(),
            escapeJson(type),
            escapeJson(event.getRealmId()),
            escapeJson(event.getClientId()),
            escapeJson(event.getUserId()),
            escapeJson(event.getIpAddress()),
            escapeJson(event.getError())
        );

        sendWebhook(jsonPayload);
    }

    @Override
    public void onEvent(AdminEvent adminEvent, boolean includeRepresentation) {
        if (adminEvent == null) return;
        
        String operationType = adminEvent.getOperationType() != null ? adminEvent.getOperationType().toString() : "UNKNOWN";
        String resourceType = adminEvent.getResourceType() != null ? adminEvent.getResourceType().toString() : "UNKNOWN";

        String jsonPayload = String.format(
            "{" +
            "\"id\":\"%s\"," +
            "\"time\":%d," +
            "\"operationType\":\"%s\"," +
            "\"resourceType\":\"%s\"," +
            "\"resourcePath\":\"%s\"," +
            "\"realmId\":\"%s\"," +
            "\"authDetails\":{" +
                "\"realmId\":\"%s\"," +
                "\"clientId\":\"%s\"," +
                "\"userId\":\"%s\"," +
                "\"ipAddress\":\"%s\"" +
            "}," +
            "\"error\":\"%s\"" +
            "}",
            escapeJson(adminEvent.getId()),
            adminEvent.getTime(),
            escapeJson(operationType),
            escapeJson(resourceType),
            escapeJson(adminEvent.getResourcePath()),
            escapeJson(adminEvent.getRealmId()),
            adminEvent.getAuthDetails() != null ? escapeJson(adminEvent.getAuthDetails().getRealmId()) : "",
            adminEvent.getAuthDetails() != null ? escapeJson(adminEvent.getAuthDetails().getClientId()) : "",
            adminEvent.getAuthDetails() != null ? escapeJson(adminEvent.getAuthDetails().getUserId()) : "",
            adminEvent.getAuthDetails() != null ? escapeJson(adminEvent.getAuthDetails().getIpAddress()) : "",
            escapeJson(adminEvent.getError())
        );

        sendWebhook(jsonPayload);
    }

    private void sendWebhook(String jsonPayload) {
        if (webhookUrl == null || webhookUrl.isEmpty()) {
            logger.fine("DevHub SPI Webhook URL is not configured. Skipping push.");
            return;
        }

        try {
            HttpRequest.Builder builder = HttpRequest.newBuilder()
                    .uri(URI.create(webhookUrl))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(3))
                    .POST(HttpRequest.BodyPublishers.ofString(jsonPayload));

            if (webhookSecret != null && !webhookSecret.isEmpty()) {
                builder.header("X-Webhook-Secret", webhookSecret);
            }

            HttpRequest request = builder.build();

            // Send asynchronously to avoid blocking Keycloak transaction
            httpClient.sendAsync(request, HttpResponse.BodyHandlers.ofString())
                    .thenAccept(response -> {
                        if (response.statusCode() != 200) {
                            logger.warning("DevHub SPI Webhook push failed. HTTP Status: " + response.statusCode());
                        }
                    })
                    .exceptionally(ex -> {
                        logger.log(Level.SEVERE, "Error pushing Keycloak event webhook to DevHub", ex);
                        return null;
                    });

        } catch (Exception e) {
            logger.log(Level.SEVERE, "Failed to build or send Keycloak event webhook", e);
        }
    }

    private String escapeJson(String str) {
        if (str == null) return "";
        return str.replace("\\", "\\\\")
                  .replace("\"", "\\\"")
                  .replace("\b", "\\b")
                  .replace("\f", "\\f")
                  .replace("\n", "\\n")
                  .replace("\r", "\\r")
                  .replace("\t", "\\t");
    }

    @Override
    public void close() {
        // No-op
    }
}
