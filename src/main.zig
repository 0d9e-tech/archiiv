const std = @import("std");
const log = std.log.scoped(.main);
const zap = @import("zap");

fn nyi(r: zap.SimpleRequest) void {
    r.sendBody("Not yet implemented") catch return;
}

const routes = std.ComptimeStringMap(zap.SimpleHttpRequestFn, .{
    .{ "/upload_file", nyi },
    .{ "/tree", nyi },
    .{ "/list", nyi },
    .{ "/list_shared_with_me", nyi },
    .{ "/get_permissions", nyi },
    .{ "/set_permissions", nyi },
});

fn dispatch(r: zap.SimpleRequest) void {
    if (r.path) |path| {
        log.info("request: {s}", .{path});
        if (routes.get(path)) |foo| {
            return foo(r);
        }
    }
    r.sendBody("Unknown endpoint") catch return;
}

pub fn main() !void {
    var listener = zap.SimpleHttpListener.init(.{
        .port = 3000,
        .on_request = dispatch,
        .log = false,
    });
    try listener.listen();

    log.info("started", .{});

    // start worker threads
    zap.start(.{
        .threads = 2,
        .workers = 2,
    });
}
