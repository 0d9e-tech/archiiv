const std = @import("std");
const log = std.log.scoped(.tree_endpoint);
const Alc = std.mem.Allocator;
const Server = std.http.Server;
const fs = std.fs;

const endpointh = @import("endpoint_helper.zig");
const bad = endpointh.bad;

pub fn handle(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) void {
    return handleInner(res, alc, root, path) catch |e| return endpointh.serverErr(e, res);
}

pub fn handleInner(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) !void {
    if (try endpointh.getUserFromHeadersLeaky(alc, res.request.headers, root)) |user| {
        var user_dir = try root.openDir(user.name, .{});
        defer user_dir.close();

        var target_dir = try user_dir.openIterableDir(
            if (path.len == 0) "./" else path,
            .{},
        );
        defer target_dir.close();

        const json = try constructJsonLeaky(alc, &target_dir);

        try res.do();
        res.transfer_encoding = .{ .content_length = json.len };
        try res.headers.append("Content-Type", "application/json");
        try res.headers.append("Connection", "close");
        try res.writeAll(json);
        try res.finish();
    } else {
        return bad(res, .unauthorized);
    }
}

fn constructJsonLeaky(alc: Alc, root: *fs.IterableDir) ![]u8 {
    var buffer = std.ArrayList(u8).init(alc);
    var jws = std.json.writeStream(buffer.writer(), .{});
    var itr = root.iterate();
    try jws.beginArray();
    while (try itr.next()) |p| {
        try jws.write(p);
    }
    try jws.endArray();
    return buffer.items;
}
