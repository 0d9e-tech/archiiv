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

see routes.go
