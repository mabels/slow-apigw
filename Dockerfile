FROM alpine:3

#RUN apk add iptables

COPY ./slow-apigw /bin/slow-apigw

ENTRYPOINT ["/bin/slow-apigw"]
