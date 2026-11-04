FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG APP=api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${APP}

FROM alpine:3.22
RUN adduser -D -h /app app
WORKDIR /app
COPY --from=build /out/app /app/app
COPY db /app/db
RUN mkdir -p /app/data && chown -R app:app /app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
