const std = @import("std");
const log = std.log.scoped(.tree_endpoint);
const fs = std.fs;
const http = std.http;

pub fn handle(user_dir: fs.Dir, path: []const u8, sink: anytype) !http.Status {
    var target_dir = user_dir.openIterableDir(
        if (path.len == 0) "./" else path,
        .{},
    ) catch |e| return switch (e) {
        error.NotDir, error.FileNotFound, error.BadPathName, error.InvalidUtf8 => .not_found,
        else => e,
    };
    defer target_dir.close();

    try constructJson(&target_dir, sink);

    return .ok;
}

fn constructJson(root: *fs.IterableDir, sink: anytype) !void {
    var jws = std.json.writeStream(sink, .{});
    var itr = root.iterate();
    try jws.beginArray();
    while (try itr.next()) |p| {
        try jws.write(p);
    }
    try jws.endArray();
}
