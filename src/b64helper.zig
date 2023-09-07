const std = @import("std");

/// Helper function that turns byte arrays into byte arrays that {de,}serialise
/// as base64 strings in the std.json functions
pub fn Base64EncDec(comptime len: comptime_int) type {
    return struct {
        value: [len]u8,

        const Self = @This();

        /// Length of the base64-encoded string
        pub const B64Len = std.base64.url_safe.Encoder.calcSize(len);

        /// Overwrite default json stringify to serialise this as base64 string instead of array
        pub fn jsonStringify(self: @This(), jws: anytype) !void {
            //jws is json.stringify.WriteStream
            var buffer: [B64Len]u8 = undefined;
            const slice = std.base64.url_safe.Encoder.encode(&buffer, &self.value);
            _ = try jws.write(slice);
        }

        /// Overwrite default json parse to deserialise this from base64 string instead of array
        pub fn jsonParse(
            alc: std.mem.Allocator,
            source: anytype,
            _: std.json.ParseOptions,
        ) !@This() {
            switch (try source.nextAlloc(alc, .alloc_if_needed)) {
                .string, .allocated_string => |str| {
                    if (str.len != Self.B64Len) {
                        return error.UnexpectedToken;
                    }
                    const sliced = str[0..Self.B64Len];
                    var self: Self = undefined;
                    std.base64.url_safe.Decoder.decode(&self.value, sliced) catch return error.UnexpectedToken;
                    return self;
                },
                else => return error.UnexpectedToken,
            }
        }

        /// Overwrite default json parse to deserialise this from base64 string instead of array
        pub fn jsonParseFromValue(
            _: std.mem.Allocator,
            src: std.json.Value,
            _: std.json.ParseOptions,
        ) !@This() {
            if (src != .string) {
                return error.UnexpectedToken;
            } else {
                return Self.decode(src.string) catch return error.UnexpectedToken;
            }
        }
    };
}

/// Serialise to base64
pub fn serialiseLeaky(value: anytype, alc: std.mem.Allocator) ![]const u8 {
    // value --stringify--> json --encode--> base64
    var list = std.ArrayListUnmanaged(u8){};
    try std.json.stringify(value, .{}, list.writer(alc));
    const len = std.base64.url_safe.Encoder.calcSize(list.items.len);
    var buffer = try alc.alloc(u8, len);
    return std.base64.url_safe.Encoder.encode(buffer, list.items);
}

/// Deserialise from base64
pub fn deserialise(comptime T: type, alc: std.mem.Allocator, str: []const u8) !?T {
    // base64 --decode--> json --parse--> T
    const len = std.base64.url_safe.Decoder.calcSizeForSlice(str) catch return null;
    var json_buffer = try alc.alloc(u8, len);
    std.base64.url_safe.Decoder.decode(json_buffer, str) catch return null;
    return std.json.parseFromSliceLeaky(T, alc, json_buffer, .{}) catch return null;
}
