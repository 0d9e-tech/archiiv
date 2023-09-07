const std = @import("std");
const log = std.log.scoped(.login_endpoint);
const fs = std.fs;
const Alc = std.mem.Allocator;
const Server = std.http.Server;

const cryptoh = @import("crypto_helper.zig");
const fsh = @import("fs_helper.zig");

const endpointh = @import("endpoint_helper.zig");
const bad = endpointh.bad;

// /whoami/

pub fn handle(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) void {
    return handleInner(res, alc, root, path) catch |e| return endpointh.serverErr(e, res);
}

pub fn handleInner(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) !void {
    if (path.len != 0) {
        return bad(res, .bad_request);
    }

    if (try endpointh.getUserFromHeadersLeaky(alc, res.request.headers, root)) |user| {
        res.transfer_encoding = .chunked;

        try res.do();

        try res.headers.append("Content-Type", "application/json");
        try res.headers.append("Connection", "close");
        // We dont send the entire User struct since it contains the otp secret.
        try std.json.stringify(.{ .name = user.name, .id = user.id }, .{}, res.writer());
        try res.finish();
    } else {
        return bad(res, .unauthorized);
    }
}
