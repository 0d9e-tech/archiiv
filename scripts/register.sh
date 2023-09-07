#!/bin/sh
set -e
zig run generate_user_secret.zig | grep "^otpauth" | qrencode -o - | feh -
echo "Now append the single entry in ./.users to archiv's .users file and change the name and id fields"
