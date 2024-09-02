FROM alpine:3.20

ARG TARGETPLATFORM

RUN apk --update add \
    ca-certificates 

RUN mkdir /app

COPY build/${TARGETPLATFORM}/moroz /app/moroz

CMD ["moroz"]
