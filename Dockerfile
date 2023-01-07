FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# run stage
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/main .
COPY .env . 
COPY start.sh . 
COPY wait-for.sh . 
COPY db/migration ./db/migration

EXPOSE 4000

CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]

