const std = @import("std");
const log = std.log.scoped(.tree_endpoint);
const Alc = std.mem.Allocator;
const fs = std.fs;

// TODO
pub fn handle(res: *std.http.Server.Response, alc: Alc, root: fs.Dir, path: []const u8) void {
    _ = res;
    _ = root;
    _ = alc;
    _ = path;
}
