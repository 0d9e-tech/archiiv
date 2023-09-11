const std = @import("std");
const log = std.log.default;
const fsh = @import("fs_helper.zig");
const loop = @import("loop.zig").loop;

pub fn main() void {
    // we are a well behaved program
    greet();
    defer farewell();

    // this is the main allocator for the whole application
    var gpa = std.heap.GeneralPurposeAllocator(.{
        .thread_safe = true,
    }){};
    defer _ = gpa.deinit();
    var alc = gpa.allocator();

    // dedicated arena for the leaky config reader
    var arena = std.heap.ArenaAllocator.init(alc);
    defer arena.deinit();
    const conf = fsh.readConfigLeaky(arena.allocator()) catch |e| {
        log.err("Failed to read config file: {}", .{e});
        return;
    };

    // the directory where archív will operate in
    const root = std.fs.openDirAbsolute(conf.root, .{}) catch |e| {
        log.err("Failed to open archív root directory '{s}': {}", .{ conf.root, e });
        return;
    };

    loop(alc, conf, root);
}

fn greet() void {
    const epoch_seconds = std.time.epoch.EpochSeconds{
        .secs = @intCast(std.time.timestamp()),
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

