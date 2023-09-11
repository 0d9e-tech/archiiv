// This module wraps crypto details and exposes simple types that nicely json
// {de,}serialise

const std = @import("std");
const testing = std.testing;
const crypto = std.crypto;
const epoch = std.time.epoch;
const HmacSha1 = crypto.auth.hmac.HmacSha1;
const Base64EncDec = @import("base64_helper.zig").Base64EncDec;

const base32 = @import("base32.zig");
const User = @import("user.zig").User;
const UserId = @import("user.zig").UserId;

// TODO sign implementation choosen arbitrarily
const sign = crypto.sign.Ed25519;

/// The secret is 256 bytes
/// Be careful:
/// - they are written to disk as base64
/// - they are send to the user as base32 (TOTP spec said so)
/// - they are written in memory as base256 bytes
pub const OTPSecret = Base64EncDec(256);

pub const Session = struct {
    // payload is _extern_ to guarantee consistent memory layout which we
    // need because we are iterating over the bytes of this when signing
    payload: extern struct {
        user_id: UserId,
        time: i64,
    },
    signature: Base64EncDec(sign.Signature.encoded_length),
};

pub const Secret = Base64EncDec(sign.SecretKey.encoded_length);

pub fn generateSecret() error{IdentityElement}!Secret {
    const key = try sign.KeyPair.create(null);
    return Secret{ .value = key.secret_key.toBytes() };
}

pub fn signSession(user_id: UserId, secret: Secret) error{ InvalidEncoding, IdentityElement, NonCanonical, KeyMismatch, WeakPublicKey }!Session {
    // this function returns empty error set for some reason..
    // we handle all 0 of those errors
    const skey = sign.SecretKey.fromBytes(secret.value) catch |e| switch (e) {};
    const key = try sign.KeyPair.fromSecretKey(skey);

    var ses = Session{
        .payload = .{
            .user_id = user_id,
            .time = std.time.timestamp(),
        },
        .signature = undefined,
    };

    const payload_bytes = std.mem.asBytes(&ses.payload);
    var noise_bytes: [sign.noise_length]u8 = undefined;
    std.crypto.random.bytes(&noise_bytes);

    const sig = try key.sign(payload_bytes, noise_bytes);
    ses.signature = .{ .value = sig.toBytes() };

    return ses;
}

pub fn verifySignedSession(secret: Secret, session: Session) error{ NonCanonical, InvalidEncoding, IdentityElement, WeakPublicKey }!?UserId {
    // this function returns empty error set for some reason..
    // we handle all 0 of those errors
    const skey = sign.SecretKey.fromBytes(secret.value) catch |e| switch (e) {};
    const pkey = sign.PublicKey.fromBytes(skey.publicKeyBytes()) catch |e| switch (e) {
        error.NonCanonical => return e,
    };
    const signature = sign.Signature.fromBytes(session.signature.value);
    const payload_bytes = std.mem.asBytes(&session.payload);
    signature.verify(payload_bytes, pkey) catch |e| switch (e) {
        error.SignatureVerificationFailed => return null,
        error.NonCanonical => return error.NonCanonical,
        error.InvalidEncoding => return error.InvalidEncoding,
        error.IdentityElement => return error.IdentityElement,
        error.WeakPublicKey => return error.WeakPublicKey,
    };
    return session.payload.user_id;
}

// Based on https://github.com/gizmo-ds/totp-wasm-zig/blob/main/src/otp.zig
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
        @as(u8, @truncate(counter >> 0)),
    };

    HmacSha1.create(hmac[0..], counter_bytes[0..], key);

    const offset = hmac[hmac.len - 1] & 0xf;
    const bin_code = hmac[offset .. offset + 4];
    const int_code =
        @as(u32, bin_code[3]) << 0 |
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
fn totp(secret: []const u8, t: i64, digit: u32, period: u32) u32 {
    const counter = @divFloor(t, period);
    return hotp(secret, @as(u64, @bitCast(counter)), digit);
}

pub fn checkOtpCodeIsValid(user: User, code: []const u8) bool {
    const time = std.time.timestamp();
    const local_code = totp(&user.otp_secret.value, time, 6, 30);
    const remote_code = std.fmt.parseInt(u32, code, 10) catch |e| {
        switch (e) {
            error.Overflow,
            error.InvalidCharacter,
            => return false,
        }
    };
    return local_code == remote_code;
}

pub fn generateOtpSecret() OTPSecret {
    var s: OTPSecret = undefined;
    std.crypto.random.bytes(&s.value);
    return s;
}

test "totp test" {
    {
        const secret = try base32.decode(testing.allocator, "GM4VC2CQN5UGS33ZJJVWYUSFMQ4HOQJW");
        defer testing.allocator.free(secret);
        try testing.expectEqual(
            @as(u32, 473526),
            totp(secret, 1662681600, 6, 30),
        );
    }
    {
        const secret = try base32.decode(testing.allocator, "3N2OTFHXKLR2E3WNZSYQ====");
        defer testing.allocator.free(secret);
        try testing.expectEqual(
            @as(u32, 29283),
            totp(secret, 1650183739, 6, 30),
        );
    }
}
