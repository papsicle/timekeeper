docker run --rm --user "$(id -u)":"$(id -g)" -v "$PWD":/usr/src/myapp -w /usr/src/myapp rust:1.39-alpine3.10 /bin/sh -c "apk add libc-dev && cargo build --release"
