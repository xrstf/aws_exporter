FROM golang:1.15-alpine as builder

WORKDIR /app/
COPY . .
RUN go build

FROM alpine:3.12

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/aws_exporter .
EXPOSE 9759
ENTRYPOINT ["/aws_exporter"]
