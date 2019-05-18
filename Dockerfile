FROM alpine:3.9 as base

RUN apk add --no-cache libraw libjpeg-turbo libexif ca-certificates openssl ffmpeg
RUN update-ca-certificates
#libexif glib expat
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --allow-untrusted vips


# ==================================

FROM base as builder

RUN apk add --no-cache ssl_client libssl1.1

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
#RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

ENV GOLANG_VERSION 1.11.4

RUN set -eux; \
	apk add --no-cache --virtual .build-deps \
		bash \
		gcc \
		musl-dev \
		openssl \
		go \
	; \
	export \
# set GOROOT_BOOTSTRAP such that we can actually build Go
		GOROOT_BOOTSTRAP="$(go env GOROOT)" \
# ... and set "cross-building" related vars to the installed system's values so that we create a build targeting the proper arch
# (for example, if our build host is GOARCH=amd64, but our build env/image is GOARCH=386, our build needs GOARCH=386)
		GOOS="$(go env GOOS)" \
		GOARCH="$(go env GOARCH)" \
		GOHOSTOS="$(go env GOHOSTOS)" \
		GOHOSTARCH="$(go env GOHOSTARCH)" \
	; \
# also explicitly set GO386 and GOARM if appropriate
# https://github.com/docker-library/golang/issues/184
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		armhf) export GOARM='6' ;; \
		x86) export GO386='387' ;; \
	esac;

RUN apk add --no-cache curl
RUN	curl -L "https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz" > go.tgz

RUN	echo '4cfd42720a6b1e79a8024895fa6607b69972e8e32446df76d6ce79801bbadb15 *go.tgz' | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	cd /usr/local/go/src; \
	./make.bash; \
	\
	rm -rf \
# https://github.com/golang/go/blob/0b30cf534a03618162d3015c8705dd2231e34703/src/cmd/dist/buildtool.go#L121-L125
		/usr/local/go/pkg/bootstrap \
# https://golang.org/cl/82095
# https://github.com/golang/build/blob/e3fe1605c30f6a3fd136b561569933312ede8782/cmd/release/releaselet.go#L56
		/usr/local/go/pkg/obj \
	; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

# ----

RUN apk --no-cache upgrade
RUN apk add --no-cache build-base libraw-dev libjpeg-turbo-dev libexif-dev git
RUN apk add --no-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --allow-untrusted vips vips-dev

COPY ./src /app/src
COPY ./vendor /app/vendor
COPY ./go.mod /app/
COPY ./go.sum /app/
COPY ./Makefile /app/

WORKDIR /app
RUN make build


# ==========================================================

FROM base as app

WORKDIR /app
COPY ./static /app/static
COPY ./public /app/public
COPY ./bin/play-to-stdout.sh /app/bin/play-to-stdout.sh
COPY ./bin/thumbnail-from-video.sh /app/bin/thumbnail-from-video.sh
COPY --from=builder /app/bin/pine-trees /app/bin/pine-trees
RUN chmod a+x /app/bin/pine-trees
RUN chmod a+x /app/bin/play-to-stdout.sh
RUN chmod a+x /app/bin/thumbnail-from-video.sh

ENV GALLERY_PATH /photos
VOLUME /photos

CMD /app/bin/pine-trees
