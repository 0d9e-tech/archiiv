// Tiny helper to generate sample .users file
// simply run with `zig run`

const std = @import("std");
const cryptoh = @import("crypto_helper.zig");
const fsh = @import("fs_helper.zig");
const User = @import("user.zig").User;

pub fn main() !void {
    const users = [_]User{
        User{
            .id = 19,
            .name = "prokop",
            .otp_secret = undefined,
        },
    };
    try fsh.writeUsers(&users, std.fs.cwd());
}
