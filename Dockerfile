# STAGE 1: build
FROM golang:1.15.2-alpine AS build

# Get dependencies
#RUN go get github.com/glavk/autoheal

WORKDIR /go/src/autoheal

# Copy all sources to workdir
COPY . .

# Setup enviromnet
ENV VERBOSE=0
ENV ARCH=amd64
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV VERSION=test

# Do the build statical link file
RUN go build -a -installsuffix cgo -o /go/bin/autoheal .


# STAGE 2: Deployment
#FROM alpine
#USER nobody:nobody
FROM scratch
COPY --from=build /go/bin/autoheal /autoheal 
CMD [ "/autoheal" ]
