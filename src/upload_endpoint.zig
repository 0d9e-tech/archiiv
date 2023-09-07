const std = @import("std");
const log = std.log.scoped(.login_endpoint);
const fs = std.fs;
const Alc = std.mem.Allocator;
const Server = std.http.Server;

const cryptoh = @import("crypto_helper.zig");
const fsh = @import("fs_helper.zig");

const endpointh = @import("endpoint_helper.zig");
const bad = endpointh.bad;

// /upload/path/to/some/dir/file
// Any directories are automatically created

pub fn handle(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) void {
    return handleInner(res, alc, root, path) catch |e| return endpointh.serverErr(e, res);
}

pub fn handleInner(res: *Server.Response, alc: Alc, root: fs.Dir, path: []const u8) !void {
    // TODO: handle thumbnails and image/video format conversion
    if (try endpointh.getUserFromHeadersLeaky(alc, res.request.headers, root)) |user| {
        const file_bytes = try res.reader().readAllAlloc(alc, std.math.maxInt(usize));

        var user_dir = try root.openDir(user.name, .{});
        defer user_dir.close();

        if (std.fs.path.dirname(path)) |dirname| {
            try user_dir.makePath(dirname);
        }

        var target_file = user_dir.createFile(path, .{ .exclusive = true }) catch |e| {
            switch (e) {
                error.PathAlreadyExists => return bad(res, .conflict),
                else => return e,
            }
        };
        defer target_file.close();

        try target_file.writer().writeAll(file_bytes);

        try res.do();
    } else {
        return bad(res, .unauthorized);
    }
}
