FROM docker:27.5-dind

RUN apk add --no-cache curl

# Install Go
ENV GOLANG_VERSION=1.23.3
RUN curl -fsSL https://golang.org/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install Air
RUN go install github.com/air-verse/air@v1.61.5
ENV PATH="/root/go/bin:${PATH}"

WORKDIR /var/app

# Trick to make use of caching
COPY go.mod go.sum /var/app/
RUN go mod download

COPY . /var/app

CMD ["air"]
