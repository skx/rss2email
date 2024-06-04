#
# Trivial Dockerfile for rss2email.
#
# Build it like so:
#
#   docker build -t rss2email:latest .
#
# I tag/push to a github repository like so:
#
#   docker tag rss2email:latest docker.pkg.github.com/skx/docker/rss2email:7
#   docker push docker.pkg.github.com/skx/docker/rss2email:7
#
# Running it will be something like this:
#
#    docker run -d \
#         --env SMTP_HOST=smtp.gmail.com \
#         --env SMTP_USERNAME=steve@example.com \
#         --env SMTP_PASSWORD=secret \
#         rss2email:latest daemon -verbose steve@example.com
#

# STEP1 - Build-image
###########################################################################
FROM golang:alpine AS builder

ARG VERSION

LABEL org.opencontainers.image.source=https://github.com/skx/rss2email/

# Create a working-directory
WORKDIR $GOPATH/src/github.com/skx/rss2email/

# Copy the source to it
COPY . .

# Build the binary - ensuring we pass the build-argument
RUN go build -ldflags "-X main.version=$VERSION" -o /go/bin/rss2email

# STEP2 - Deploy-image
###########################################################################
FROM alpine

# Copy the binary.
COPY --from=builder /go/bin/rss2email /usr/local/bin/

# Set entrypoint
ENTRYPOINT [ "/usr/local/bin/rss2email" ]

# Set default command
CMD help

# Create a group and user
RUN addgroup app && adduser -D -G app -h /app app

# Tell docker that all future commands should run as the app user
USER app

# Set working directory
WORKDIR /app
