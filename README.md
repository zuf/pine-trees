# Pine Trees

**STATUS: this project in very early stage and not ready for everyday usage**

Pine Trees is a fast web gallery for viewing photos from RAW files.

Main purpose: browse your huge photo RAW archive at your NAS / home or office server.


## Benchmarks

At this experimental stage for raw files from the 7D DSLR (18 megapixels CR2) on i7 2600 CPU (linux x86_64):

- fetch full frame preview (jpeg) from raw file over http - about 30-40ms
- get thumbnail (jpeg 300x200) from raw (in runtime) - about 100-200ms

That speed allow you to browse your photo archive without any precalculated caches and offline thumbnail generators.


## External dependencies:

 - [libraw](https://www.libraw.org/)
 - [libvips](https://libvips.github.io/libvips/) (through [bimg](https://github.com/h2non/bimg))
 - [ffmpeg](https://ffmpeg.org/) (for video remuxing on the fly)
 
 
 ### Run in Docker

Change `/your_photos` to path to directory with your photos. 
 
`docker run -v /your_photos:/photos:ro -p 1323:1323 -e GALLERY_PATH=/photos zufzzi/pine-trees`
 
Open in browser: http://localhost:1323/
 
Or use `docker-compose.yml` file lake that:
 
 ```yaml
version: '3.6'

services

  pine-trees:
    image: zufzzi/pine-trees    
    ports:
      - 1323:1323
    environment:
      - GALLERY_PATH=/photos
    volumes:
      - /path/to/your/photos:/photos:ro
```

