FROM golang:alpine as builder

LABEL Maintainer=deanmax@gmail.com

# Note: make sure you run docker build from root of repo
COPY . /src
WORKDIR /src/cmd/wuqumo

RUN go build -o /build/wuqumo

# hadolint ignore=DL3006
FROM alpine:3.17 as runner

ARG user=app
ARG group=app
ARG base=/app

RUN mkdir "${base}"
WORKDIR /app

COPY --from=builder /build .

RUN addgroup "${group}" && \
  adduser -D -u 1000 -G "${group}" "${user}" && \
  chown -Rv "${user}:${group}" "${base}"

USER ${user}
CMD ["./wuqumo"]

EXPOSE 8000
