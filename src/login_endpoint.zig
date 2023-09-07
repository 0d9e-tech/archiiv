const std = @import("std");
const log = std.log.scoped(.login_endpoint);
const mem = std.mem;
const fs = std.fs;
const Alc = mem.Allocator;
const Server = std.http.Server;

const fsh = @import("fs_helper.zig");

const cryptoh = @import("crypto_helper.zig");
const Session = cryptoh.Session;

const endpointh = @import("endpoint_helper.zig");
const bad = endpointh.bad;

// /login/:username/:otpCode/

pub fn handle(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) void {
    return handleInner(res, alc, root, path) catch |e| return endpointh.serverErr(e, res);
}

pub fn handleInner(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) !void {
    // '/login/' was stripped
    var itr = mem.splitScalar(u8, path, '/');
    const username: []const u8 = itr.next() orelse return bad(res, .bad_request);
    const otp_code: []const u8 = itr.next() orelse return bad(res, .bad_request);

    const user = try endpointh.getUserFromUsernameLeaky(alc, root, username) orelse return bad(res, .unauthorized);

    if (!try cryptoh.validateOtpCode(user, otp_code)) {
        return bad(res, .unauthorized);
    }

    const secret = try fsh.getSecretLeaky(alc, root);
    const new_session = try cryptoh.signSession(user.id, secret);
    try endpointh.setSessionCookieLeaky(alc, &res.headers, new_session);
    try res.do();

    log.info("user '{s}' logged in", .{username});
}
