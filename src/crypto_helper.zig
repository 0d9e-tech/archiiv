// This module wraps crypto details and exposes simple types that nicely json
// {de,}serialise

const std = @import("std");
const crypto = std.crypto;
const epoch = std.time.epoch;
const Base64EncDec = @import("b64helper.zig").Base64EncDec;

const User = @import("user.zig").User;
const UserId = @import("user.zig").UserId;

// TODO choosen arbitrarily
const sign = crypto.sign.Ed25519;

/// The secret is 256 bytes
/// Note: they are written to disk as base64
/// Note: the TOTP spec works with base32
/// Care is needed when converting this
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

pub fn generateSecret() !Secret {
    const key = try sign.KeyPair.create(null);
    return Secret{ .value = key.secret_key.toBytes() };
}

pub fn signSession(user_id: UserId, secret: Secret) !Session {
    const skey = try sign.SecretKey.fromBytes(secret.value);
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

pub fn verifySignedSession(secret: Secret, session: Session) ?UserId {
    const skey = sign.SecretKey.fromBytes(secret.value) catch return null;
    const pkey = sign.PublicKey.fromBytes(skey.publicKeyBytes()) catch return null;
    const signature = sign.Signature.fromBytes(session.signature.value);
    const payload_bytes = std.mem.asBytes(&session.payload);
    signature.verify(payload_bytes, pkey) catch return null;
    return session.payload.user_id;
}
