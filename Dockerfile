FROM golang:1.6.3
RUN go get -v github.com/reyoung/GhIssueBot
CMD "GhIssueBot" "-c" "/config.yaml"

