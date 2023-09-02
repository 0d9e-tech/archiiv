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
const Session = @import("crypto_helper.zig").Session;
const b64h = @import("b64helper.zig");

// inlined to catch the error return trace from caller
pub inline fn serverErr(err: anyerror, res: *Server.Response) void {
    log.err("unhandled server error: {s}", .{@errorName(err)});
    if (@errorReturnTrace()) |trace| { // logs error return trace if we have debug info
        std.debug.dumpStackTrace(trace.*);
    }
    return bad(res, .internal_server_error);
}

pub fn bad(res: *Server.Response, status: http.Status) void {
    // prints helpful stack trace for debugging but spams the log a lot
    //log.debug("request declared bad", .{});
    //std.debug.dumpCurrentStackTrace(null);

    res.status = status;
    res.do() catch |e| {
        log.err("Failed to send http response headers: {}", .{e});
    };
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

pub fn setSessionCookieLeaky(alc: Alc, headers: *http.Headers, session: Session) !void {
    var cookie_header = std.ArrayListUnmanaged(u8){};
    const writer = cookie_header.writer(alc);
    _ = try writer.write("archiv-session-secret=");

    const session_base64 = try b64h.serialiseLeaky(session, alc);
    _ = try writer.write(session_base64);

    _ = try writer.write("; Secure; Path=/");
    return headers.append("Set-Cookie", cookie_header.items);
}
