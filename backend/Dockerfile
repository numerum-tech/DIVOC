FROM golang:1.15-alpine AS build_base

RUN  apk add make git gcc musl-dev
# Set the Current Working Directory inside the container
WORKDIR /tmp/divoc
COPY vaccination_api/go.mod ./vaccination_api/go.mod
COPY vaccination_api/go.sum ./vaccination_api/go.sum
COPY kernel_library/go.mod ./kernel_library/go.mod
COPY kernel_library/go.sum ./kernel_library/go.sum
COPY portal_api/go.mod ./portal_api/go.mod
COPY portal_api/go.sum ./portal_api/go.sum
COPY notification-service/go.mod ./notification-service/go.mod
COPY notification-service/go.sum ./notification-service/go.sum

RUN cd vaccination_api && go mod download
RUN cd portal_api && go mod download
RUN cd notification-service && go mod download
COPY vaccination_api/Makefile ./vaccination_api/Makefile
RUN cd vaccination_api && make deps
COPY . .
RUN find .
RUN cd vaccination_api && GOFLAGS=" -tags=musl" SPEC_FILE="../dockerdeps/interfaces/vaccination-api.yaml" make all
RUN cd portal_api && GOFLAGS=" -tags=musl" SPEC_FILE="../dockerdeps/interfaces/admin-portal.yaml" make all
RUN cd notification-service && GOFLAGS=" -tags=musl" SPEC_FILE="../dockerdeps/interfaces/notification-service.yaml" make all

# Start fresh from a smaller image
FROM alpine:3.9 as vaccination_api
LABEL name=vaccination_api
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/vaccination_api/divoc-server /app/divoc-server
COPY vaccination_api/config /config
EXPOSE 8000
#RUN useradd -u 8000 vaccination_api
USER 8000
CMD ["/app/divoc-server", "--scheme", "http", "--port", "8000", "--host", "0.0.0.0"]

FROM alpine:3.9 as analytics_feed
LABEL name=analytics_feed
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/vaccination_api/analytics_feed /app/analytics_feed
COPY vaccination_api/config /config
#RUN useradd -u 1001 analytics_feed
USER 1001
CMD ["/app/analytics_feed"]

FROM alpine:3.9 as certificate_processor
LABEL name=certificate_processor
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/vaccination_api/certificate_processor /app/certificate_processor
COPY vaccination_api/config /config
#RUN useradd -u 1001 certificate_processor
USER 1002
CMD ["/app/certificate_processor"]

#portal api
FROM alpine:3.9 as portal_api
LABEL name=portal_api
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/portal_api/divoc-portal-api-server /app/divoc-portal-api-server
COPY portal_api/config /config
EXPOSE 8001
#RUN useradd -u 8001 portal_api
USER 8001
CMD ["/app/divoc-portal-api-server", "--scheme", "http", "--port", "8001", "--host", "0.0.0.0"]

#notification-service
FROM alpine:3.9 as notification-service
LABEL name=notification-service
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/notification-service/notification-service /app/notification-service
COPY notification-service/config /config
EXPOSE 8765
#RUN useradd -u 8765 divoc-notification
USER 8765
CMD ["/app/notification-service", "--scheme", "http", "--port", "8765", "--host", "0.0.0.0"]

#digilocker_support_api
FROM alpine:3.9 as digilocker_support_api
LABEL name=digilocker_support_api
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/vaccination_api/digilocker_support_api /app/digilocker_support_api
COPY vaccination_api/config /config
#RUN useradd -u 8003 divoc-cert
USER 8003
CMD ["/app/digilocker_support_api"]

#caching-dash-server
FROM alpine:3.9 as caching-dash-server
LABEL name=caching-dash-server
RUN apk add ca-certificates
COPY --from=build_base /tmp/divoc/vaccination_api/caching-dash-server /app/caching-dash-server
COPY vaccination_api/config /config
#RUN useradd -u 8004 divoc-cache
USER 8004
CMD ["/app/caching-dash-server"]
