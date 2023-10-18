FROM golang as build
WORKDIR /usr/local/src/freeze-calendar-resource
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/local/bin/freeze-calendar-resource -ldflags '-extldflags "-static"'

FROM registry.access.redhat.com/ubi8-minimal:latest
RUN mkdir -p /opt/resource
COPY --from=build /usr/local/bin/freeze-calendar-resource /usr/local/bin/
RUN    printf '#!/usr/bin/env bash\n/usr/local/bin/freeze-calendar-resource check "$@"' > /opt/resource/check \
    && printf '#!/usr/bin/env bash\n/usr/local/bin/freeze-calendar-resource put "$@"' > /opt/resource/out \
    && printf '#!/usr/bin/env bash\n/usr/local/bin/freeze-calendar-resource get "$@"' > /opt/resource/in \
    && chmod +x /opt/resource/*
