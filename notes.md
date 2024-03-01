## File Storage

Archív has it's own fs implemented on top of the system's fs.

files dont have a content in a traditional sense. instead, each file has
multiple separate contents called sections. The tradition main content is in the
'data' section. Metadata is stored in the 'meta' section, hooks can create own
sections for example to store image thumbnails or image exif metadata

For implementation details see the big comment in fs/fs.go

## Sharing

### UX

To share a file, modify the file permissions to be accessible to the target
users and then share the file uuid through some external channel to the target
user.

The target user then mounts this uuid in their directory of choice.

This will not create any new copies of the shared file. If someone you shared
the file with has write permissions and edits the file, others will see the
changes too. If you want to revoke the file share, you can just remove the read
and write permissions from the people you don't want to have the file.

Sharing with other people can be only done by people with the `owner` permission
bit.

## Permissions

Permissions are specified for each user with these three bits:

- read - read the file
- write - write the file
- owner - write the file's metadata

These permissions are not inherited through the filesystem, but are set for each
file separately. However Archív offers an API to quickly set permission bits for
a file tree. There is a special user called `pub`, who anyone can be logged in
as. Another special user is `root`, who has access to anything, but can't be
logged in as. If a user doesn't have any permissions specified for a file, they
have the same permissions as the `pub` user. Archív offers the ability to create
groups of users.

## Metadata

see fs/filemeta.go

## Upload Hooks

Archiiv triggers file hooks when file is uploaded/deleted/edited. Hooks can be
enabled by globs or per file.

Directory hooks are triggered when a file in the is uploaded into/deleted from
the directory.

Hooks can modify the uploaded file, create/modify/delete other files/directories
and use external programs to do so.

Hook ideas:

| Hook name  | Description                                                              |
| ---------- | ------------------------------------------------------------------------ |
| Exif       | Extracts exif metadata from the file and puts it into the metadata json. |
| Thumbnails | Creates thumbnails from the files.                                       |
| Archiver   | Backups the file or directory in a compressed archive.                   |
| Exec       | Executes an external process.                                            |
| ImgConvert | Converts uploaded image into a sane format and compression quality       |

## API

see routes.go
