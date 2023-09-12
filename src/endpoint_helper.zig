const std = @import("std");
const log = std.log.scoped(.endpoint_helpers);
const http = std.http;
const Server = http.Server;
const mem = std.mem;
const fs = std.fs;
const Alc = mem.Allocator;
const user = @import("user.zig");
const User = user.User;
const UserId = user.UserId;
const fsh = @import("fs_helper.zig");
const cryptoh = @import("crypto_helper.zig");
const Session = cryptoh.Session;
const b64h = @import("base64_helper.zig");

// inlined to catch the error return trace from caller
pub inline fn serverErr(res: *Server.Response, err: anyerror) void {
    log.err("unhandled server error: {s}", .{@errorName(err)});
    if (@errorReturnTrace()) |trace| { // logs error return trace if we have debug info
        std.debug.dumpStackTrace(trace.*);
    }
    return setStatus(res, .internal_server_error);
}

pub fn setStatus(res: *Server.Response, status: http.Status) void {
    // prints helpful stack trace for debugging but spams the log a lot
    //log.debug("request declared bad", .{});
    //std.debug.dumpCurrentStackTrace(null);

    res.status = status;
    res.do() catch |e| {
        log.err("Failed to send http response headers: {}", .{e});
    };
}

// TODO: instead of null return public user
pub fn getUserFromHeadersLeaky(alc: Alc, headers: http.Headers, root: fs.Dir) !?User {
    const session = try getSessionCookie(alc, headers) orelse return null;
    const secret = try fsh.getSecretLeaky(alc, root);
    if (try cryptoh.verifySignedSession(secret, session)) |user_id| {
        return try getUserFromUserIdLeaky(alc, root, user_id);
    }
    return null;
}

pub fn getUserFromUsernameLeaky(alc: Alc, root: fs.Dir, username: []const u8) !?User {
    const users = try fsh.getUsersLeaky(alc, root);
    for (users) |u| {
        if (mem.eql(u8, u.name, username)) {
            return u;
        }
    }
    return null;
}

pub fn getUserFromUserIdLeaky(alc: Alc, root: fs.Dir, userId: UserId) !?User {
    const users = try fsh.getUsersLeaky(alc, root);
    for (users) |u| {
        if (u.id == userId) {
            return u;
        }
    }
    return null;
}

pub fn getSessionCookie(alc: Alc, headers: http.Headers) !?Session {
    const all_cookies = blk: {
        for (headers.list.items) |header| {
            if (std.ascii.eqlIgnoreCase(header.name, "cookie")) {
                break :blk header.value;
            }
        }
        return null;
    };

    const our_cookie = blk: {
        var iter = mem.tokenizeSequence(u8, all_cookies, ", ");
        while (iter.next()) |c| {
            if (mem.startsWith(u8, c, "archiv-session-secret=")) {
                break :blk c;
            }
        }
        return null;
    };

    const our_cookie_value = our_cookie[22..];

    return b64h.deserialise(Session, alc, our_cookie_value);
}

pub fn openUserDirectory(root: fs.Dir, name: []const u8) !fs.Dir {
    return try root.openDir(name, .{});
}

pub fn setSessionCookieLeaky(alc: Alc, headers: *http.Headers, session: Session) !void {
    var cookie_header = std.ArrayListUnmanaged(u8){};
    const session_base64 = try b64h.serialiseLeaky(session, alc);

    const writer = cookie_header.writer(alc);
    try writer.print("archiv-session-secret={s}; Secure; Path=/", .{session_base64});

    return headers.append("Set-Cookie", cookie_header.items);
}

pub fn validatePath(alc: Alc, path: []const u8) Alc.Error!?[]u8 {
    const resolved_path = try std.fs.path.resolvePosix(alc, &[_][]const u8{path});

    // forbid absolute paths or paths that travel up the dir tree
    if (resolved_path[0] == '/' or resolved_path[0] == '.') {
        return null;
    }
    return resolved_path;
}
