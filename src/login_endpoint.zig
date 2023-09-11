const std = @import("std");
const log = std.log.scoped(.login_endpoint);
const mem = std.mem;
const fs = std.fs;
const http = std.http;
const Alc = mem.Allocator;
const Server = std.http.Server;

const fsh = @import("fs_helper.zig");
const cryptoh = @import("crypto_helper.zig");
const endpointh = @import("endpoint_helper.zig");

// /login/:username/:otpCode/

pub fn handle(headers: *http.Headers, alc: Alc, root: fs.Dir, path: []const u8) !http.Status {
    var itr = mem.splitScalar(u8, path, '/');
    const username: []const u8 = itr.next() orelse return .bad_request;
    const otp_code: []const u8 = itr.next() orelse return .bad_request;

    const user = try endpointh.getUserFromUsernameLeaky(alc, root, username) orelse return .unauthorized;

    if (!try cryptoh.validateOtpCode(user, otp_code)) {
        return .unauthorized;
    }

    const secret = try fsh.getSecretLeaky(alc, root);
    const new_session = try cryptoh.signSession(user.id, secret);
    try endpointh.setSessionCookieLeaky(alc, headers, new_session);

    log.info("user '{s}' logged in", .{username});

    return .ok;
}
