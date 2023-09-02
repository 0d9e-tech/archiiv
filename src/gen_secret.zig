// Tiny helper to generate server secrets
// simply run with `zig run`

const std = @import("std");
const cryptoh = @import("crypto_helper.zig");
const fsh = @import("fs_helper.zig");

pub fn main() !void {
    const secret = try cryptoh.generateSecret();
    try fsh.writeSecret(secret, std.fs.cwd());
}
