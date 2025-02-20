# Use an official Golang image as the base image
FROM golang:1.23-alpine as stage
# update packages if any
RUN apk update

# Create build image for go
FROM stage as build

# Set the Go environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install necessary tools
RUN apk update && apk add --no-cache make
# Set the working directory inside the container
WORKDIR /app
# Copy the entire project into the container
COPY . .
# Ensure Go binaries are in PATH
#ENV PATH="/go/bin:${PATH}"

# Verify installations
RUN go version

# Build the project digital-core
RUN make build

# discard build image and use fresh image for copying binary
FROM stage AS final
WORKDIR /app
COPY --from=build /app/bin/goapps /app/bin/goapps

EXPOSE 8080
ENTRYPOINT ["/app/bin/goapps"]