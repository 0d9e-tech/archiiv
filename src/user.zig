const std = @import("std");
const cryptoh = @import("crypto_helper.zig");

pub const UserId = u64;

pub const User = struct {
    id: UserId,
    name: []const u8,
    otp_secret: cryptoh.OTPSecret,
};

