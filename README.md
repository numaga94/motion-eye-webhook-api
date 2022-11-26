# motion-eye-webhook-api

## env configs

    PORT=4000
    TOKEN="YOUR TELEGRAM BOT TOKEN"
    CHAT_ID="YOUR CHAT ID"
    SNAPSHOT_URL="YOUR CAMERA SHOT URL"
    SWITCH_URL="YOUR DEVICE LOCAL IP"
    AUTH_KEY="YOUR API AUTH_KEY"

## build for raspberry pi zero: ARMv6

    `env GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags="-s -w" -o main-ARMv6 .`
