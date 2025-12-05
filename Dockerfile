FROM golang:alpine
MAINTAINER MKP Moble Production <mkpproduction@gmail.com>

# Set necessary environmet variables needed for our image
#ENV GO111MODULE=on \
#    CGO_ENABLED=0 \
#    GOOS=linux \
#    GOARCH=amd64

## Create an /data directory within our
RUN mkdir /data

## Copy everything in the root directory
## into our /data directory

ADD . /data

## Specify that we now wish to execute
## any further commands inside our /data
## directory

WORKDIR /data

# EXPOSED Specify volumes on project
#VOLUME /data/assets

COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

EXPOSE 8080

## Our start command which kicks off
## our newly created binary executable
CMD ["sample"]