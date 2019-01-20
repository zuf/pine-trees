# Pine Trees

**STATUS: this project in very early stage and not ready for everyday usage**

Pine Trees is a fast web gallery for viewing photos from RAW files.

Main purpose: browse your huge photo RAW archive at your NAS / home or office server.


## Benchmarks

At this experimental stage for raw files from the 7D DSLR (18 megapixels CR2) on i7 2600 CPU (linux x86_64):

- fetch full frame preview (jpeg) from raw file over http - about 30-40ms
- get thumbnail (jpeg 300x200) from raw (in runtime) - about 100-200ms

That speed allow you to browse your photo archive without any precalculated caches and offline thumbnail generators.


### External dependencies:

 - [libraw](https://www.libraw.org/)
 - [libvips](https://libvips.github.io/libvips/) (through [bimg](https://github.com/h2non/bimg))
 