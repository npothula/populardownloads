# Build stage of assetcounter at intermediate container
FROM golang:1.14.1-alpine3.11 AS dev_img

ARG HOME_PATH="/home/jfrog-test"
ARG APP_PATH="jfrog-test/svcPopularDownloads"
ARG GO_APP_PATH="/go/src/jfrog-test/svcPopularDownloads"
ARG APP_BIN_PATH="/home/jfrog-test/bin/svcPopularDownloads"

RUN apk update && apk upgrade \
    && apk add --no-cache ca-certificates

RUN mkdir -p $HOME_PATH/log
RUN mkdir -p $HOME_PATH/tmp

# copying the source code and build in the container
RUN mkdir -p $GO_APP_PATH
WORKDIR $GO_APP_PATH
COPY . $GO_APP_PATH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -o $APP_BIN_PATH $GO_APP_PATH/src/cmd/
RUN chmod +x $APP_BIN_PATH


# final stage
FROM alpine:3.11 AS prod_img

RUN mkdir -p /home/jfrog-test \
    && apk --no-cache add ca-certificates

WORKDIR /home/jfrog-test
COPY --from=dev_img /home/jfrog-test .

CMD ["/home/jfrog-test/bin/svcPopularDownloads"]
