const std = @import("std");
const log = std.log.scoped(.login_endpoint);
const http = std.http;
const User = @import("user.zig").User;

// /whoami/

pub fn handle(user: User, path: []const u8, sink: anytype) !http.Status {
    if (path.len != 0) {
        return .bad_request;
    }

    // We dont send the entire User struct since it contains the otp secret.
    try std.json.stringify(.{ .name = user.name, .id = user.id }, .{}, sink);

    return .ok;
}
