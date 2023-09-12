const std = @import("std");
const mem = std.mem;
const assert = std.debug.assert;
const fs = std.fs;
const log = std.log.scoped(.loop);
const Alc = mem.Allocator;
const Thread = std.Thread;
const Server = std.http.Server;

const Config = @import("Config.zig");

const endpointh = @import("endpoint_helper.zig");
const setStatus = endpointh.setStatus;

const lsEndpoint = @import("ls_endpoint.zig").handle;
const loginEndpoint = @import("login_endpoint.zig").handle;
const whoamiEndpoint = @import("whoami_endpoint.zig").handle;
const uploadEndpoint = @import("upload_endpoint.zig").handle;

pub fn loop(alc: Alc, conf: Config, root: fs.Dir) void {
    var server = Server.init(alc, .{
        .reuse_port = true,
        .reuse_address = true,
    });
    defer server.deinit();

    const address = std.net.Address.parseIp("127.0.0.1", conf.port) catch |e| {
        log.err("Failed to parse IP: {}", .{e});
        return;
    };

    server.listen(address) catch |e| {
        log.err("Failed to listen: {}", .{e});
        return;
    };

    var pool: Thread.Pool = undefined;
    Thread.Pool.init(&pool, .{ .allocator = alc }) catch |e| {
        log.err("Failed to start up thread pool: {}", .{e});
        return;
    };
    defer pool.deinit();

    while (true) {
        var res = server.accept(.{ .allocator = alc }) catch |e| {
            log.err("Failed to accept connection (ignoring it): {}", .{e});
            continue;
        };

        pool.spawn(handle, .{ res, alc, root }) catch |e| {
            log.err("Failed to spawn a task in the pool to handle incoming connection (ignoring it): {}", .{e});
            continue;
        };
    }
}

fn handle(res_: Server.Response, main_alc: Alc, root: fs.Dir) void {
    // local mutable copy
    var res = res_;

    // response is not allocated in the arena and deinit also closes the
    // connection so we really have to call it
    defer res.deinit();

    // We wrap the global allocator in a arena that is destroyed at the end of
    // the request. The arena is used as a fallback when the stack allocator
    // runs out.
    var arena = std.heap.ArenaAllocator.init(main_alc);
    defer {
        log.debug("Arena has {} nodes after request", .{arena.state.buffer_list.len()});
        arena.deinit();
    }
    // 4k choosen arbitrarily
    var stack_alc = std.heap.stackFallback(4096, arena.allocator());
    const alc = stack_alc.get();

    res.wait() catch |e| {
        log.err("Failed to wait for the response: {}", .{e});
        return;
    };

    dispatchToHandler(alc, res, root) catch |e| return endpointh.serverErr(&res, e);
}

fn dispatchToHandler(alc: Alc, res_: Server.Response, root: fs.Dir) !void {
    // local copy; res is in the waited state
    var res = res_;
    assert(res.state == .waited);
    var path = stripQueryAndFragment(res.request.target);

    var path_iter = mem.splitScalar(u8, path[1..], '/');
    const endpoint_name = path_iter.next() orelse return setStatus(&res, .bad_request);
    const path_rest = path_iter.rest();

    log.info("Handling request to: {s}", .{endpoint_name});

    // login is the only endpoint that doesn't need logged in user
    if (mem.eql(u8, endpoint_name, "login")) {
        const status = try loginEndpoint(&res.headers, alc, root, path_rest);
        return setStatus(&res, status);
    }

    const user = try endpointh.getUserFromHeadersLeaky(alc, res.request.headers, root) orelse return setStatus(&res, .unauthorized);
    const user_dir = try endpointh.openUserDirectory(root, user.name);

    // upload endpoint is the only one that reads from the request
    if (mem.eql(u8, endpoint_name, "upload")) {
        const payload = try res.reader().readAllAlloc(alc, std.math.maxInt(usize));
        const status = try uploadEndpoint(alc, user_dir, path_rest, payload);
        return setStatus(&res, status);
    }

    res.transfer_encoding = .chunked;

    // temp buffer for the json response
    var response_buffer = std.ArrayListUnmanaged(u8){};
    var writer = response_buffer.writer(alc);

    // call the handler
    const status = if (mem.eql(u8, endpoint_name, "ls"))
        try lsEndpoint(alc, user_dir, path_rest, writer)
    else if (mem.eql(u8, endpoint_name, "whoami"))
        try whoamiEndpoint(user, path_rest, writer)
    else
        .not_found;

    setStatus(&res, status);
    assert(res.state == .responded);
    try res.writeAll(response_buffer.items);
    try res.finish();
}

fn stripQueryAndFragment(path_: []const u8) []const u8 {
    var path = path_;
    // Strip fragment (can it even be here?)
    if (mem.indexOfScalar(u8, path, '#')) |fragment_start| {
        path = path[0..fragment_start];
    }

    // Strip query
    if (mem.indexOfScalar(u8, path, '?')) |query_start| {
        path = path[0..query_start];
    }

    return path;
}

// TODO missing endpoints:
// - cat
// - cat_thumbnail
// - cat_meta? - for photo location and time info? what about nonphotos
//   - bulk version?
// - some directory meta data getter and setter
//   - owner, permissions&sharing, thumbnail options
//   - inheritance
// - share some dir with someone
// - delete/trash file?
// - move file?
