const std = @import("std");
const Alc = std.mem.Allocator;
const log = std.log.scoped(.login_endpoint);
const fs = std.fs;
const http = std.http;
const endpointh = @import("endpoint_helper.zig");

// /upload/path/to/some/dir/file
// Any directories are automatically created

pub fn handle(alc: Alc, user_dir: fs.Dir, dangerous_path: []const u8, payload: []const u8) !http.Status {
    // TODO: handle thumbnails and image/video format conversion
    const path = try endpointh.validatePath(alc, dangerous_path) orelse return .forbidden;

    if (fs.path.dirname(path)) |dirname| {
        try user_dir.makePath(dirname);
    }

    var target_file = user_dir.createFile(path, .{ .exclusive = true }) catch |e| {
        switch (e) {
            error.PathAlreadyExists => return .conflict,
            else => return e,
        }
    };
    defer target_file.close();
    try target_file.writer().writeAll(payload);

    return .ok;
}
