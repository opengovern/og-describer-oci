FROM docker.io/golang:alpine as build
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./local/og-describer-oci ./
ENTRYPOINT [ "./og-describer-oci" ]
CMD [ "./og-describer-oci" ]