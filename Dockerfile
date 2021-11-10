FROM golang:1.16-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go get -u -v -f all

COPY . .

RUN go build -o ./pardugo ./main.go

FROM alpine:3.13 as prod

WORKDIR /app

COPY --from=build /app/pardugo ./pardugo

CMD [ "./pardugo"]
