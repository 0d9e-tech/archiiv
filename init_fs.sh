#/bin/sh
set -xeu

# usage: ./init_fs.sh FS_ROOT_PATH

mkdir $1
uuid=`uuidgen`
touch $1/$uuid
echo '{"is_dir":true,"name":"root"}' > $1/$uuid
echo "fs root is $uuid"
