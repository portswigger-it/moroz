FROM ubuntu:24.04

ARG TARGETPLATFORM

# COPY ./build/linux/moroz /usr/bin/moroz

RUN apt-get update \
    && apt-get install -y ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app

COPY build/${TARGETPLATFORM}/moroz /moroz

RUN /moroz --version

CMD [./moroz]
