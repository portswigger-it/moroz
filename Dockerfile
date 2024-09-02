FROM alpine:3.20

RUN apk --update add \
    ca-certificates 

COPY ./build/linux/moroz /usr/bin/moroz

CMD ["moroz"]
