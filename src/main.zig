const std = @import("std");
const mem = std.mem;
const time = std.time;
const assert = std.debug.assert;
const log = std.log.default;
const fs = std.fs;
const Alc = mem.Allocator;
const Server = std.http.Server;
const Thread = std.Thread;
const fsh = @import("fs_helper.zig");
const Config = @import("Config.zig");

const lsEndpoint = @import("ls_endpoint.zig").handle;
const loginEndpoint = @import("login_endpoint.zig").handle;
const whoamiEndpoint = @import("whoami_endpoint.zig").handle;

pub fn main() void {
    // we are a well behaved program
    greet();
    defer farewell();

    // this is the main allocator for the whole application
    var gpa = std.heap.GeneralPurposeAllocator(.{
        .thread_safe = true,
        .stack_trace_frames = 0,
    }){};
    defer _ = gpa.deinit();
    var alc = gpa.allocator();

    var arena = std.heap.ArenaAllocator.init(alc);
    defer arena.deinit();
    const conf = fsh.readConfigLeaky(arena.allocator()) catch |e| {
        log.err("Failed to read config file: {}", .{e});
        return;
    };

    // The directory where archív will operate in
    const root = fs.openDirAbsolute(conf.root, .{}) catch |e| {
        log.err("Failed to open archív root directory '{s}': {}", .{ conf.root, e });
        return;
    };

    loop(alc, conf, root);
}

fn greet() void {
    const epoch_seconds = time.epoch.EpochSeconds{
        .secs = @intCast(time.timestamp()),
    };
    const day_seconds = epoch_seconds.getDaySeconds();
    const hour_of_day = day_seconds.getHoursIntoDay();
    log.info("Good {s}", .{switch (hour_of_day) {
        0...11 => "morning",
        12...19 => "afternoon",
        else => "evening",
    }});
}

fn farewell() void {
    log.info("Farewell", .{});
}

fn loop(alc: Alc, conf: Config, root: fs.Dir) void {
    var s = Server.init(alc, .{
        .reuse_port = true,
        .reuse_address = true,
    });
    defer s.deinit();

    const address = std.net.Address.parseIp("127.0.0.1", conf.port) catch unreachable;
    s.listen(address) catch |e| {
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
        var res = s.accept(.{ .allocator = alc }) catch |e| {
            log.err("Failed to accept connection (ignoring): {}", .{e});
            continue;
        };

        pool.spawn(handle, .{ res, alc, root }) catch |e| {
            log.err("Failed to spawn a thread to handle incoming connection (ignoring): {}", .{e});
            continue;
        };
    }
}

const endpoints = std.ComptimeStringMap(*const fn (*Server.Response, Alc, fs.Dir, []const u8) void, .{
    .{ "/ls/", lsEndpoint },
    .{ "/login/", loginEndpoint },
    .{ "/whoami/", whoamiEndpoint },
    //.{ "/", ... },
    //.{ "/upload/", ... },
    //.{ "/lsshared/", ... },
    //.{ "/getperm/", ... },
    //.{ "/setperm/", ... },
    // ...
});

fn handle(res_: Server.Response, main_alc: Alc, root: fs.Dir) void {
    var res = res_; // local mutable copy

    // response is not allocated in the arena
    defer res.deinit();

    // We use arena allocator for the entire request and then throw the arena
    // away at the end
    var arena = std.heap.ArenaAllocator.init(main_alc);
    defer {
        log.debug("Alloc list has {} nodes after request", .{arena.state.buffer_list.len()});
        arena.deinit();
    }
    const alc = arena.allocator();

    res.wait() catch |e| {
        log.err("Failed to wait for the response: {}", .{e});
        return;
    };

    var path = res.request.target;

    // Strip fragment (can it even be here?)
    if (mem.indexOfScalar(u8, path, '#')) |fragment_start| {
        path = path[0..fragment_start];
    }

    // Strip query
    if (mem.indexOfScalar(u8, path, '?')) |querry_start| {
        path = path[0..querry_start];
    }

    const pos = mem.indexOfScalarPos(u8, path, 1, '/') orelse 0;
    // Extract the endpoint name
    const first_segment = path[0 .. pos + 1];

    log.info("Handling request to: {s}", .{first_segment});

    // dispatch to endpoint handlers
    if (endpoints.get(first_segment)) |handler| {
        handler(&res, alc, root, path[pos + 1 ..]);
    } else {
        res.status = .not_found;
        res.do() catch |e| {
            log.err("Failed to send headers: {}", .{e});
        };
    }
}
