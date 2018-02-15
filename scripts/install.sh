apk --update upgrade
apk add --no-cache \
    bash make curl git \
    ca-certificates build-base \
    libxml2-dev protobuf \
    yarn

make prepare-build
make packages
