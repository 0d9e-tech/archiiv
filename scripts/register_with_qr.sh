#!/bin/sh
zig run register_user.zig 2>&1 | grep "^otpauth" | qrencode -o - | feh - &
echo "Now append the single entry in ./.users to archiv's .users file"
