# parts:

* server
* mobile app
* web view

## Mobile app

Single task:

* upload of local files to server using webdav

## Server

External webdav with multiple users

## Web view

* fowards login info to webdav server
  * webdav server can be running on different server
* generates everything on demand
  * thumbnails cached
  * html not cached
* has to support most image formats
  * (we dont have a deamon that converts whatever was uploaded to something
    sane)

## anything else

just make a standalone thing that manipulates the filesystem and has its own
interface.

* url to upload things
  * could literally just be [gimmedat](https://github.com/vakabus/gimmedat)
  * needs webdav fs mounted but we are going to use this rarely
* google photos importer
  * just some script that is run manually
* tool that optimizes random formats of questionable compression ratio uploaded
  from a phone
  * just some shell script (webview can display the format in any case) (gnu
    parallel + imagemagick/ffmpeg)
  * can be handled manually when disk usage becomes a problem

