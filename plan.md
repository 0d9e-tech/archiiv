# parts:

* server
* mobile app
* web view

## Mobile app

**Single** task:

* upload of local files to server using **already existing** secure-ftp-style
  protocol
  * webdav? <- this might be the thing
  * smb?
  * ftps?
  * sftp?

## Server

**already existing** secure-ftp-style server

## Web view

* pigallery style thing
* generates everything on demand
  * (thumbnails cached)
  * avoids problems with having to refresh when files are uploaded (we dont need
    a daemon)
* has to support most image formats
  * (we dont have a deamon that converts whatever was uploaded to something
    sane)

## anything else

just make a standalone thing that manipulates the filesystem and has its own
interface.

* url to upload things
  * separate web service that uploads into some directory
  * could literally just be [gimmedat](https://github.com/vakabus/gimmedat)
* google photos importer
  * just some script that is run manually
* tool that optimizes random formats of questionable compression ratio uploaded
  from a phone
  * just some script that is run manually (webview can display the format in any
    case) (gnu parallel + imagemagick/ffmpeg)
  * can be handled manually when disk usage becomes a problem

