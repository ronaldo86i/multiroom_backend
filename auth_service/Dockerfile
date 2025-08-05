# Etapa de construcción
FROM golang:1.24.5-alpine AS build

# Instalar dependencias necesarias
RUN apk add --no-cache git tzdata

# Establecer la variable de entorno TZ
ENV TZ=America/La_Paz

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app
COPY .env go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o multiroom_auth ./cmd

FROM alpine:latest

ENV TZ=America/La_Paz

RUN apk add --no-cache ca-certificates tzdata

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=build /app/multiroom_auth .

EXPOSE 8080

CMD ["/app/multiroom_auth"]
