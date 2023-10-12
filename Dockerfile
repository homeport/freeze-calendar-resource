FROM registry.access.redhat.com/ubi8/go-toolset:latest AS build
COPY . .
RUN go build -buildvcs=false -o freeze-calendar-resource .

FROM registry.access.redhat.com/ubi8-minimal:latest
RUN mkdir -p /opt/resource
COPY --from=build /opt/app-root/src/freeze-calendar-resource /opt/resource
RUN    printf '#!/usr/bin/env bash\n/opt/resource/freeze-calendar-resource check "$@"' > /opt/resource/check \
    && printf '#!/usr/bin/env bash\n/opt/resource/freeze-calendar-resource put "$@"' > /opt/resource/out \
    && printf '#!/usr/bin/env bash\n/opt/resource/freeze-calendar-resource get "$@"' > /opt/resource/in \
    && chmod +x /opt/resource/*
