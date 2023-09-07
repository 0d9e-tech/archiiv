// Based on https://github.com/gizmo-ds/totp-wasm-zig/blob/main/src/base32.zig
const std = @import("std");
const testing = std.testing;
const Alc = std.mem.Allocator;

const RFC4648_ALPHABET: *const [32]u8 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

pub fn encode(alc: Alc, input: []const u8, padding: bool) ![]u8 {
    var output = try std.ArrayList(u8).initCapacity(alc, (input.len + 3) / 4 * 5);
    defer output.deinit();

    var itr = std.mem.window(u8, input, 5, 5);
    while (itr.next()) |chunk| {
        var buf = [_]u8{0} ** 5;
        for (chunk, 0..) |b, i| {
            buf[i] = b;
        }

        try output.append(RFC4648_ALPHABET[(buf[0] & 0xF8) >> 3]);
        try output.append(RFC4648_ALPHABET[((buf[0] & 0x07) << 2) | ((buf[1] & 0xC0) >> 6)]);
        try output.append(RFC4648_ALPHABET[(buf[1] & 0x3E) >> 1]);
        try output.append(RFC4648_ALPHABET[((buf[1] & 0x01) << 4) | ((buf[2] & 0xF0) >> 4)]);
        try output.append(RFC4648_ALPHABET[(buf[2] & 0x0F) << 1 | (buf[3] >> 7)]);
        try output.append(RFC4648_ALPHABET[(buf[3] & 0x7C) >> 2]);
        try output.append(RFC4648_ALPHABET[((buf[3] & 0x03) << 3) | ((buf[4] & 0xE0) >> 5)]);
        try output.append(RFC4648_ALPHABET[buf[4] & 0x1F]);
    }

    if (input.len % 5 != 0) {
        const len = output.items.len;
        const num_extra = 8 - (input.len % 5 * 8 + 4) / 5;
        if (padding) {
            for (1..num_extra + 1) |i| {
                output.items[len - i] = '=';
            }
        } else {
            try output.resize(len - num_extra);
        }
    }
    return output.toOwnedSlice();
}

pub fn decode(alc: Alc, input: []const u8) ![]u8 {
    var unpad = input.len;
    for (1..@min(6, input.len) + 1) |i| {
        if (input[input.len - i] != '=') break;
        unpad -= 1;
    }

    const output_len = unpad * 5 / 8;

    var output = try std.ArrayList(u8).initCapacity(alc, (output_len + 4) / 5 * 5);
    defer output.deinit();

    var itr = std.mem.window(u8, input, 8, 8);
    while (itr.next()) |chunk| {
        var buf = [_]u8{0} ** 8;
        for (chunk, 0..) |b, ci| {
            if (std.mem.indexOf(u8, RFC4648_ALPHABET, &[1]u8{b})) |v| {
                buf[ci] = @as(u8, @intCast(v));
            }
        }

        try output.append((buf[0] << 3) | (buf[1] >> 2));
        try output.append((buf[1] << 6) | (buf[2] << 1) | (buf[3] >> 4));
        try output.append((buf[3] << 4) | (buf[4] >> 1));
        try output.append((buf[4] << 7) | (buf[5] << 2) | (buf[6] >> 3));
        try output.append((buf[6] << 5) | (buf[7]));
    }
    try output.resize(output_len);
    return output.toOwnedSlice();
}

test "base32 encode test" {
    const alc = testing.allocator;

    var output = try encode(alc, "Hello world", true);
    defer alc.free(output);

    try testing.expectEqualSlices(u8, "JBSWY3DPEB3W64TMMQ======", output);
}

test "base32 decode test" {
    const alc = testing.allocator;

    var output = try decode(alc, "JBSWY3DPEB3W64TMMQ======");
    defer alc.free(output);

    try testing.expectEqualSlices(u8, "Hello world", output);
}
