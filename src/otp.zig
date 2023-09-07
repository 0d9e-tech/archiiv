// Based on https://github.com/gizmo-ds/totp-wasm-zig/blob/main/src/otp.zig
const std = @import("std");
const testing = std.testing;
const Alc = std.mem.Allocator;
const HmacSha1 = std.crypto.auth.hmac.HmacSha1;
const User = @import("user.zig").User;

const cryptoh = @import("crypto_helper.zig");
const base32 = @import("base32.zig");

fn hotp(key: []const u8, counter: u64, digit: u32) u32 {
    var hmac: [HmacSha1.mac_length]u8 = undefined;
    const counter_bytes = [8]u8{
        @as(u8, @truncate(counter >> 56)),
        @as(u8, @truncate(counter >> 48)),
        @as(u8, @truncate(counter >> 40)),
        @as(u8, @truncate(counter >> 32)),
        @as(u8, @truncate(counter >> 24)),
        @as(u8, @truncate(counter >> 16)),
        @as(u8, @truncate(counter >> 8)),
        @as(u8, @truncate(counter)),
    };

    HmacSha1.create(hmac[0..], counter_bytes[0..], key);

    const offset = hmac[hmac.len - 1] & 0xf;
    const bin_code = hmac[offset .. offset + 4];
    const int_code =
        @as(u32, bin_code[3]) |
        @as(u32, bin_code[2]) << 8 |
        @as(u32, bin_code[1]) << 16 |
        @as(u32, bin_code[0]) << 24 & 0x7FFFFFFF;

    const code = int_code % (std.math.pow(u32, 10, digit));
    return code;
}

test "hotp test" {
    try testing.expectEqual(
        @as(u32, 886679),
        hotp("GM4VC2CQN5UGS33ZJJVWYUSFMQ4HOQJW", 1662681600, 6),
    );
}

/// Note: secret is already base32 decoded here
fn totp(secret: []const u8, t: i64, digit: u32, period: u32) !u32 {
    const counter = @divFloor(t, period);
    return hotp(secret, @as(u64, @bitCast(counter)), digit);
}

pub fn validateOtpCode(user: User, code: []const u8) !bool {
    const time = std.time.timestamp();
    const local_code = try totp(&user.otp_secret.value, time, 6, 30);
    const remote_code = std.fmt.parseInt(u32, code, 10) catch |e| {
        switch (e) {
            error.Overflow,
            error.InvalidCharacter,
            => return false,
        }
    };
    return local_code == remote_code;
}

pub fn generateOtpSecret() cryptoh.OTPSecret {
    var s: cryptoh.OTPSecret = undefined;
    std.crypto.random.bytes(&s.value);
    return s;
}

test "totp test" {
    {
        const secret = try base32.decode(testing.allocator, "GM4VC2CQN5UGS33ZJJVWYUSFMQ4HOQJW");
        defer testing.allocator.free(secret);
        try testing.expectEqual(
            @as(u32, 473526),
            try totp(secret, 1662681600, 6, 30),
        );
    }
    {
        const secret = try base32.decode(testing.allocator, "3N2OTFHXKLR2E3WNZSYQ====");
        defer testing.allocator.free(secret);
        try testing.expectEqual(
            @as(u32, 29283),
            try totp(secret, 1650183739, 6, 30),
        );
    }
}
