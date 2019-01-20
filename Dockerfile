FROM golang:1.11-alpine as builder
RUN apk add --no-cache build-base libjpeg-turbo libjpeg-turbo-dev libexif libexif-dev libraw libraw-dev wget git tar glib glib-dev  expat expat-dev

RUN mkdir -p /vips
WORKDIR /vips
RUN wget https://github.com/libvips/libvips/releases/download/v8.7.4/vips-8.7.4.tar.gz
RUN tar xzf vips-8.7.4.tar.gz
WORKDIR /vips/vips-8.7.4
RUN ./configure
RUN make -j8
RUN make install

COPY ./ /app
WORKDIR /app
RUN make build

#FROM alpine as app
#RUN apk add --no-cache libjpeg-turbo libraw libexif glib expat
#
#COPY --from=builder /vips/ /vips/
#WORKDIR /vips/vips-8.7.4
#RUN ls -lh
#RUN apk add --no-cache make
#RUN make install
#RUN rf -rf /vips
#
#COPY --from=builder /app/bin/pine-tree /app/
#WORKDIR /app

VOLUME /photos
ENV GALLERY_PATH /photos

CMD /app/bin/pine-trees

