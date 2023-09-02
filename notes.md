
---

Prokop misc notes

# server:

* bez database??
  * dirs, files
  * */.meta/        * meta info o fotkách (lokace, čas, ..?)
  * */.thumb/       * thumbnails
  * */.permissions  * (edit, add, delete, ..?) per user
  * /.users         * list users, hash hesla?
* link na
  * photo dump do složky
  * photo access
* shared photos between dirs?? symlink?? nebo není potřeba?
* jxl + on demand convert do jpg or smt pro web?

# mobile app:

* upload fotky z množiny složek (může být na button)
* mapování local dir -> remote dir
  * měl by umět vyrábět složky po měsících týden, ...
    * format string setting?
* smazat remotly backed up photos

# misc:

bulk import z google photos tool!!

---

Odsouhlaseno na meetingu:

* configy na disku jsou json
* přes api se posílá json

# server je API

* endpointy
  * uploadnout soubor do složky
    * target dir
    * soubor name
  * list dir tree
  * list files
  * list shared with me (slow probably)
  * get dir permissions
  * set dir permissions

* public fake user

* README per dir

* .users file
  * jenom otp haha?
  * login s username+otp
    * dostane session token [který jdou mazat per device]--(prokop:asi ne)

* user má dir
* per directory inherited .settings

# frontend používá API
# mobil je taky frontend simple

* API do filesystemu

