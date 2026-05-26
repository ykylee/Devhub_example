package com.devhub.keycloak.spi;

import org.keycloak.Config;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;
import java.util.logging.Logger;

public class DevHubEventListenerProviderFactory implements EventListenerProviderFactory {

    private static final Logger logger = Logger.getLogger(DevHubEventListenerProviderFactory.class.getName());
    private static final String PROVIDER_ID = "devhub-event-listener";

    private String webhookUrl;
    private String webhookSecret;

    @Override
    public EventListenerProvider create(KeycloakSession session) {
        return new DevHubEventListenerProvider(webhookUrl, webhookSecret);
    }

    @Override
    public void init(Config.Scope config) {
        // Resolve from environment variables (cloud-native approach)
        webhookUrl = System.getenv("DEVHUB_BACKEND_SPI_WEBHOOK_URL");
        webhookSecret = System.getenv("DEVHUB_KEYCLOAK_SPI_WEBHOOK_SECRET");
        
        logger.info("Initializing DevHub Event Listener SPI...");
        if (webhookUrl == null || webhookUrl.isEmpty()) {
            logger.warning("DEVHUB_BACKEND_SPI_WEBHOOK_URL environment variable is not set. Real-time push will be disabled.");
        } else {
            logger.info("DevHub Event Listener SPI Webhook URL configured: " + webhookUrl);
        }
    }

    @Override
    public void postInit(KeycloakSessionFactory factory) {
        // No-op
    }

    @Override
    public void close() {
        // No-op
    }

    @Override
    public String getId() {
        return PROVIDER_ID;
    }
}
