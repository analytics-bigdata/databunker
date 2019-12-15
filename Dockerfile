############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git gcc libc-dev
RUN go get -u github.com/fatih/structs
RUN go get -u github.com/gobuffalo/packr
RUN go get -u github.com/gobuffalo/packr/packr
RUN go get -u github.com/tidwall/gjson
RUN go get -u github.com/ttacon/libphonenumber
RUN go get -u github.com/hashicorp/go-uuid
RUN go get -u go.mongodb.org/mongo-driver/bson
RUN go get -u modernc.org/ql/ql
RUN go get -u github.com/evanphx/json-patch
RUN go get -u github.com/julienschmidt/httprouter
WORKDIR $GOPATH/src/paranoidguy/databunker/src/
COPY . $GOPATH/src/paranoidguy/databunker/
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# prepare web to go with packr
RUN packr
# debug
RUN find $GOPATH/src/paranoidguy/databunker/
# Build the binary.
RUN go build -o /go/bin/databunker
# clean packr
RUN packr clean
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /bin/busybox /bin/busybox
COPY --from=builder /bin/busybox /bin/sh
COPY --from=builder /lib/ld* /lib/
COPY --from=builder /go/bin/databunker /databunker/bin/databunker
COPY run.sh /databunker/bin/
#COPY create-test-user.sh /databunker/bin/
COPY databunker.yaml /databunker/conf/
RUN /bin/busybox mkdir -p /databunker/data
RUN /bin/busybox mkdir -p /databunker/certs
#RUN /bin/busybox ln -s /bin/busybox /bin/sh
# Run the hello binary.
#ENTRYPOINT ["/go/bin/databunker"]
EXPOSE 3000
ENTRYPOINT ["/bin/sh", "/databunker/bin/run.sh"]
#CMD ["/bin/sh", "-x", "-c", "/go/bin/databunker -init"]
