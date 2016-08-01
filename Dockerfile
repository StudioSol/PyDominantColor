FROM golang:1.6-wheezy

RUN apt-get update -y && \
    apt-get install -y -qq --no-install-recommends \
    pkg-config \
    build-essential \
    python \
    python-dev \
    && rm -rf /var/lib/apt/lists/*

ENV PKG_CONFIG_PATH=/usr/lib/pkgconfig:/usr/local/lib/x86_64-linux-gnu/pkgconfig:/usr/local/lib/pkgconfig:/usr/local/share/pkgconfig:/usr/lib/x86_64-linux-gnu/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig
