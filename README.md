# Hochschule Burgenland - BSWE - Game Director

[![](https://img.shields.io/github/license/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](LICENSE.md)
[![](https://img.shields.io/github/actions/workflow/status/muhlba91/hochschule-burgenland-bswe-director/verify.yml?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/actions/workflows/verify.yml)
[![](https://img.shields.io/coverallsCoverage/github/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/)
[![](https://api.scorecard.dev/projects/github.com/muhlba91/hochschule-burgenland-bswe-director/badge?style=for-the-badge)](https://scorecard.dev/viewer/?uri=github.com/muhlba91/hochschule-burgenland-bswe-director)
[![](https://img.shields.io/github/release-date/muhlba91/hochschule-burgenland-bswe-director?style=for-the-badge)](https://github.com/muhlba91/hochschule-burgenland-bswe-director/releases)
[![](https://img.shields.io/github/all-contributors/muhlba91/hochschule-burgenland-bswe-director?color=ee8449&style=for-the-badge)](#contributors)
<a href="https://www.buymeacoffee.com/muhlba91" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="28" width="150"></a>

Game Director is a Go-based service designed to manage game sessions.

//FIXME: update readme

---

## Features

- **AI-Powered OCR**: Uses Google Gemini (e.g., `gemini-1.5-flash`) to interpret water meter readings from images.
- **MQTT Integration**: Subscribes to an image topic and publishes the processed readings.
- **Home Assistant Discovery**: Automatically creates a sensor in Home Assistant for easy monitoring.
- **Cloud Storage Backup**: Optionally uploads processed images to Scaleway Object Storage (S3 compatible).
- **Health Monitoring**: Includes a `healthz` server for liveness, readiness, and startup checks.

---

## Configuration

Configure the application using the following environment variables:

| Variable                             | Description                                         | Default                         |
| ------------------------------------ | --------------------------------------------------- | ------------------------------- |
| `METER_ID`                           | Unique identifier for the meter.                    | `water-meter`                   |
| `METER_NAME`                         | Display name for the meter.                         | `Water Meter`                   |
| `METER_MODEL`                        | Model description of the meter.                     | `ESP32 Water Meter`             |
| `BROKER_ADDRESS`                     | MQTT broker address (e.g., `tcp://localhost:1883`). | `tcp://localhost:1883`          |
| `BROKER_TOPIC_SUBSCRIPTION_TEMPLATE` | Template for image subscription topic.              | `tele/%s/image`                 |
| `BROKER_TOPIC_PUBLISH_TEMPLATE`      | Template for usage publication topic.               | `stat/%s/water/usage/state`     |
| `BROKER_CLIENT_ID`                   | MQTT client ID.                                     | *(optional)*                    |
| `BROKER_USERNAME`                    | MQTT username.                                      | *(optional)*                    |
| `BROKER_PASSWORD`                    | MQTT password.                                      | *(optional)*                    |
| `GEMINI_API_KEY`                     | Google Gemini API key.                              | *(required)*                    |
| `GEMINI_MODEL`                       | Gemini model to use.                                | `gemini-3.1-flash-lite-preview` |
| `SCW_REGION`                         | Scaleway region for S3 backup.                      | `fr-par`                        |
| `SCW_ACCESS_KEY`                     | Scaleway access key.                                | *(optional)*                    |
| `SCW_SECRET_KEY`                     | Scaleway secret key.                                | *(optional)*                    |
| `SCW_BUCKET`                         | Scaleway S3 bucket name.                            | *(optional)*                    |
| `SCW_BUCKET_PATH`                    | Path template within the bucket.                    | `watermeter/%s/`                |
| `HEALTHZ_HOST`                       | Host for the health server.                         | `0.0.0.0`                       |
| `HEALTHZ_PORT`                       | Port for the health server.                         | `8080`                          |

---

## Deployment

### Docker Run

To run the processor using Docker, you need a Google Gemini API key and an MQTT broker.

```shell
docker run -d \
  --name hochschule-burgenland-bswe-director \
  -e BROKER_ADDRESS="tcp://mqtt-broker:1883" \
  -e GEMINI_API_KEY="your-gemini-api-key" \
  -e METER_ID="my-water-meter" \
  ghcr.io/muhlba91/hochschule-burgenland-bswe-director:latest
```

---

## Testing

//TODO: Add testing instructions here.
