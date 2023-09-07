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

    const session = try endpointh.getSessionCookie(alc, res.request.headers) orelse return bad(res, .unauthorized);

    const secret = try fsh.getSecretLeaky(alc, root);

    if (cryptoh.verifySignedSession(secret, session)) |user_id| {
        const user = try endpointh.getUserFromUserIdLeaky(alc, root, user_id) orelse return bad(res, .unauthorized);

        try res.do();

        // We dont send the entire User struct since it contains the otp secret.
        // {"id":19,"name":""}
        res.transfer_encoding = .{ .content_length = 19 + user.name.len };
        try res.headers.append("Content-Type", "application/json");
        try res.headers.append("Connection", "close");
        _ = try res.writeAll("{\"id\":19,\"name\":\"");
        _ = try res.writeAll(user.name);
        _ = try res.writeAll("\"}");
        try res.finish();
    } else {
        return bad(res, .unauthorized);
    }
}
