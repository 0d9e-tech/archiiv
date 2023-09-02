const std = @import("std");
const log = std.log.scoped(.fs);

const user = @import("user.zig");
const User = user.User;

const Config = @import("Config.zig");
const Secret = @import("crypto_helper.zig").Secret;

/// Finds and parses the json config file.
pub fn readConfigLeaky(alc: std.mem.Allocator) !Config {
    // config path is either the first argument or default
    const conf_path = blk: {
        var itr = std.process.args();
        if (itr.skip()) {
            if (itr.next()) |arg| {
                break :blk arg;
            }
        }
        break :blk null;
    } orelse "/etc/archiv.json";

    log.debug("Reading config file @ '{s}'", .{conf_path});

    const cwd = std.fs.cwd();
    return readFileLeaky(Config, alc, cwd, conf_path);
}

pub fn getUsersLeaky(alc: std.mem.Allocator, root: std.fs.Dir) ![]const User {
    return readFileLeaky([]User, alc, root, ".users");
}

pub fn writeUsers(users: []const User, root: std.fs.Dir) !void {
    return writeFile(users, root, ".users");
}

pub fn getSecretLeaky(alc: std.mem.Allocator, root: std.fs.Dir) !Secret {
    return readFileLeaky(Secret, alc, root, ".secret");
}

pub fn writeSecret(secret: Secret, root: std.fs.Dir) !void {
    return writeFile(secret, root, ".secret");
}

// TODO: file locking on linux is apparently unreliable:
// https://www.kernel.org/doc/Documentation/filesystems/mandatory-locking.txt
// switch to mutexes or maybe get away with only using the atomic io operations

/// Reads @T from @absolute_path
fn readFileLeaky(
    comptime T: type,
    alc: std.mem.Allocator,
    dir: std.fs.Dir,
    path: []const u8,
) !T {
    log.debug("Reading {s} from '{s}'", .{ @typeName(T), path });

    const file = try dir.openFile(path, .{ .mode = .read_only, .lock = .shared });
    defer file.close();
    const reader = file.reader();
    var json_reader = std.json.Reader(4096, @TypeOf(reader)).init(alc, reader);
    return std.json.parseFromTokenSourceLeaky(T, alc, &json_reader, .{});
}

/// Writes @value to @absolute_path
fn writeFile(
    value: anytype,
    dir: std.fs.Dir,
    path: []const u8,
) !void {
    log.debug("Writing {s} to '{s}'", .{ @typeName(@TypeOf(value)), path });
    const file = try dir.createFile(path, .{ .lock = .exclusive });
    defer file.close();
    try std.json.stringify(value, .{ .whitespace = .indent_tab }, file.writer());
}
