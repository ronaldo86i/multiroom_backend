# Etapa de construcción
FROM golang:1.24.5-alpine AS build

# Instalar dependencias necesarias
RUN apk add --no-cache git tzdata

# Establecer la variable de entorno TZ
ENV TZ=America/La_Paz

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o multiroom_sucursal ./cmd

FROM alpine:latest

ENV TZ=America/La_Paz

RUN apk add --no-cache ca-certificates tzdata

RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=build /app/multiroom_sucursal .

EXPOSE 8083
EXPOSE 8084

CMD ["/app/multiroom_sucursal"]
