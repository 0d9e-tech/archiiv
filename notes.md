## File Storage

Files are stored in a flat structure in the `/data` directory. Since each file
has an UUID, the file can be named using that UUID. The metadata is stored as
`$UUID.json`. Some hooks may have output files. These can be stored in the
`/data` directory, but they must start with the UUID of the associated file.

The structure of the file tree is stored separately. Currently, for development
purposes, it will be stored in a simple JSON dictionary in `/tree.json`.

```
{
    UUID of a directory: [
        list of UUIDS in the directory
    ]
}
```

## Sharing

### UX

If you want to select the file, the client will allow you to choose what
people/groups you want to share it with. Then you will get the file's UUID or a
link, which you can share with the people. When the people receive your UUID,
they can enter it into their archív client and add the file to any location
they choose.

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

Permissions are specified in file's metadata. They are inherited through the
file system. There are four permission bits:

- read
- write
- metadata read
- metadata write

There is a special user called `pub`, who anyone can be logged in as. Another
special user is `root`, who has access to anything, but can't be logged in as.

## Metadata

```
{
  "perms": {
    "username": bit field with permissions
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

Hook name  | Description
-----------|--------------------------------------------------------------------------
Exif       | Extracts exif metadata from the file and puts it into the metadata json.
Thumbnails | Creates thumbnails from the files.
Archiver   | Backups the file or directory in a compressed archive.
Exec       | Executes an external process.

## API

TODO: figure out auth, I've marked some endpoints with auth. These endpoints
will pass the request to some auth function, which returns some stuff.

### POST(auth) /api/v1/fs/rm

Remove a file or directory. This action requires the user to have rw permissions
for the file or the directory and all of its contents recursively.

Input:

```
{ "path": path to file to delete }
```

### POST(auth) /api/v1/fs/mkdir

Make a directory. The directory will have the author's permissions set to 0xff.
This call requires the user to have write permissions for the parent directory.

Input:

```
{ "path": path to directory to create }
```

***

# server:

- bez database??
  - dirs, files
  - /.users \* list users, hash hesla?
- link na
  - photo dump do složky
  - photo access
- shared photos between dirs?? symlink?? nebo není potřeba?
- jxl + on demand convert do jpg or smt pro web?

# mobile app:

- upload fotky z množiny složek (může být na button)
- mapování local dir -> remote dir
  - měl by umět vyrábět složky po měsících týden, ...
    - format string setting?
- smazat remotly backed up photos

# misc:

bulk import z google photos tool!!

---

Odsouhlaseno na meetingu:

- configy na disku jsou json
- přes api se posílá json

# server je API

- endpointy

  - uploadnout soubor do složky
    - target dir
    - soubor name
  - list dir tree
  - list files
  - list shared with me (slow probably)
  - get dir permissions
  - set dir permissions

- public fake user

- README per dir

- .users file

  - jenom otp haha?
  - login s username+otp
    - dostane session token [který jdou mazat per device]--(prokop:asi ne)

- user má dir
- per directory inherited .settings

# frontend používá API

# mobil je taky frontend simple

- API do filesystemu
