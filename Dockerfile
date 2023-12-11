# Start from the official Golang base image for the build stage
FROM golang:1.21.4-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN GOOS=linux GOARCH=amd64 go build -o main cmd/main.go

# Start from scratch for a small, secure final image
FROM scratch

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

ENV EDGE_EXT_EXTRACTOR_ID = "edge-extractor"
ENV EDGE_EXT_CDF_PROJECT_NAME = "my-project"
ENV EDGE_EXT_CDF_CLUSTER = "westeurope-1"
ENV EDGE_EXT_AD_TENANT_ID = "my-tenant-id"
ENV EDGE_EXT_AD_AUTH_TOKEN_URL = "https://login.microsoftonline.com/my-tenant-id/oauth2/token"
ENV EDGE_EXT_AD_CLIENT_ID = "my-client-id"
ENV EDGE_EXT_AD_SECRET = "my-secret"
ENV EDGE_EXT_AD_SCOPES = "https://management.azure.com/.default"
ENV EDGE_EXT_CDF_DATASET_ID = "my-dataset-id"
ENV EDGE_EXT_ENABLED_INTEGRATIONS = "ip_cams_to_cdf"
ENV EDGE_EXT_LOG_LEVEL = "debug"
ENV EDGE_EXT_IS_ENCRYPTED = "false"

# Command to run the executable
CMD ["./main"]

