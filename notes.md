## File Storage

Archív has it's own fs, which stores **records** in a single folder in a flat
structure. Each record has it's own UUID, which matches the Archív file it
stores.

For each record, there is a file named by it's UUID, which stores a JSON
describing two things: the record's name and a list of UUIDs. These are UUIDs of
the records **mounted** to the record.

Each record can also have **sections**. Those are stored as separate files in
the format `$UUID.$RECORD_NAME`.

Records are reference counted. If the count reaches zero, the record and all the
sections are removed from the drive.

## Sharing

### UX

If you want to select the file, the client will allow you to choose what
people/groups you want to share it with. Then you will get the file's UUID or a
link, which you can share with the people. When the people receive your UUID,
they can enter it into their archív client and add the file to any location they
choose.

This will not create any new copies of the shared file. If someone you shared
the file with has write permissions and edits the file, others will see the
changes too. If you want to revoke the file share, you can just remove the read
and write permissions from the people you don't want to have the file. The
archív server will recognize this action and remove the file link from the
user's directory.

Sharing with other people can be only done by people with the `mw` permission
bit. Giving this to someone else essentially gives them the ownership of the
file.

### Implementation

To share files, archív will implement a feature to link to a file by UUID. To
see how this feature is implemented, refer to the File Storage section of this
document.

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

```
{
  "uuid": back link to the fs record,
  "type": MIME type of the data record,
  "perms": {
    "username": bit field with permissions,
    ...
  },
  "hooks": list of required hooks,
  "createdBy": username of creator,
  "createdAt": time of creation,
}
```

## Upload Hooks

Archiiv triggers file hooks when file is uploaded/deleted/edited. Hooks can be
enabled by globs or per file.

Directory hooks are triggered when a file in the is uploaded into/deleted from
the directory.

Archiiv offers upload hooks functionality, which run some code on the uploaded
files. They can be enabled and configured in the config json. Hooks can either
be ran for a file glob, or they can be specifically requested in the metadata
json. In case of directories, the hooks are ran whenever a file is uploaded to
the directory.

Hook ideas:

| Hook name  | Description                                                              |
| ---------- | ------------------------------------------------------------------------ |
| Exif       | Extracts exif metadata from the file and puts it into the metadata json. |
| Thumbnails | Creates thumbnails from the files.                                       |
| Archiver   | Backups the file or directory in a compressed archive.                   |
| Exec       | Executes an external process.                                            |

## API

TODO: figure out auth, I've marked some endpoints with auth. These endpoints
will pass the request to some auth function, which returns some stuff.

### Error responses

If the request fails, the server will respond with a JSON looking like this:

```json
{
  "ok": false,
  "error": "error message"
}
```

### GET(AUTH) /api/v1/fs/:uuid/ls

Lists all the mounted files in the file specified by `uuid`. Read permissions
for the file are required.

Returns:

```json
{
  "ok": true,
  "data": [
    list of UUID strings
  ]
}
```

### POST(AUTH) /api/v1/fs/:uuid/touch/:name

Creates a new file with name `name` and mounts it to `uuid`. It is not possible to
create an orphan file.

Returns:

```json
{
  "ok": true,
  "data": uuid of the new file
}
```

### POST(AUTH) /api/v1/fs/:uuid/mount/:uuid2

Mounts `uuid2` to `uuid`.

### POST(AUTH) /api/v1/fs/:uuid/unmount/:uuid2

Unmounts `uuid2` from `uuid`.

### GET(AUTH) /api/v1/fs/:uuid/section/:section

Gets the data of a file's section inside the response body.

### POST(AUTH) /api/v1/fs/:uuid/section/:section

Sets the data of a files's section. If the file doesn't have a section
of that name, it will be created.
