// Tiny helper to generate TOTP secret
// use the 'register.sh' shell script

const std = @import("std");
const cryptoh = @import("src/crypto_helper.zig");
const fsh = @import("src/fs_helper.zig");
const User = @import("src/user.zig").User;
const base32 = @import("src/base32.zig");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    var alc = gpa.allocator();

    const user = User{
        .id = 19,
        .name = "prokop",
        .otp_secret = cryptoh.generateOtpSecret(),
    };
    // We store the secret in memory as raw bytes, save it to disk as base64
    // and (as the TOTP spec dictates) send to user as base32
    const b32secret = try base32.encode(alc, &user.otp_secret.value, false);
    defer alc.free(b32secret);

    const writer = std.io.getStdOut().writer();
    try writer.print("otpauth://totp/archiv:{s}?secret={s}&issuer=archiv\n", .{ user.name, b32secret });
    try fsh.writeUsers(&[_]User{user}, std.fs.cwd());
}
