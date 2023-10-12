FROM registry.access.redhat.com/ubi8/go-toolset:latest AS build
COPY . .
RUN go build -buildvcs=false -o freeze-calendar-resource .

FROM registry.access.redhat.com/ubi8-minimal:latest
COPY --from=build /opt/app-root/src/freeze-calendar-resource /
RUN mkdir -p /opt/resource \
    && printf '#!/usr/bin/env sh\n/freeze-calendar-resource check' > /opt/resource/check \
    && printf '#!/usr/bin/env sh\n/freeze-calendar-resource put' > /opt/resource/out \
    && printf '#!/usr/bin/env sh\n/freeze-calendar-resource get' > /opt/resource/in \
    && chmod +x /opt/resource/*
