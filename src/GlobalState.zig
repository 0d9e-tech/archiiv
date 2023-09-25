const std = @import("std");
const json = std.json;
const User = @import("user.zig").User;
const UserId = @import("user.zig").UserId;
const Secret = @import("crypto_helper.zig").Secret;
const Config = @import("Config.zig");
const cryptoh = @import("crypto_helper.zig");

const Self = @This();

pub fn init(conf: Config) Self {
    _ = conf;
    var s = Self{
        ._server_secret = unreachable,
        ._users = unreachable,
        ._root = unreachable,
    };
    _ = s;
}

pub fn deinit(self: Self) void {
    _ = self;
}

pub fn getUserByName(self: *const Self, name: []const u8) ?User {
    self._users_lock.lockShared();
    defer self._users_lock.unlockShared();
    for (self._users) |user| {
        if (std.mem.eql(u8, user.name, name)) {
            return user;
        }
    }
    return null;
}

pub fn getUserById(self: *const Self, id: UserId) ?User {
    self._users_lock.lockShared();
    defer self._users_lock.unlockShared();
    for (self._users) |user| {
        if (user.id == id) {
            return user;
        }
    }
    return null;
}

pub fn createUser(self: *Self, name: []const u8) User {
    self._users_lock.lock();
    defer self._users_lock.unlock();

    const max_id: UserId = blk: {
        var t: UserId = 0;
        for (self._users) |user| t = @max(t, user.id);
        break :blk t;
    };

    const new_user = User{
        .id = max_id + 1,
        .name = name,
        .otp_secret = cryptoh.generateOtpSecret(),
    };

    // grow users by one
    try self._users.arena.allocator().realloc(
        self._users.value,
        self._users.value.len + 1,
    );
    self._users.value[self._users.value.len - 1] = new_user;

    try storeUsers(self._root, self._users.value);

    return new_user;
}

const users_file_name = ".users";

// Stores users to file
fn storeUsers(root: std.fs.Dir, users: []const User) !void {
    var file = try root.createFile(users_file_name, .{});
    defer file.close();
    try std.json.stringify(users, .{}, file.writer());
}

// Loads users from file
fn loadUsers(alc: std.mem.Allocator, root: std.fs.Dir) !json.Parsed([]User) {
    var file = try root.openFile(users_file_name, .{});
    defer file.close();
    var file_reader = file.reader();
    // 256 choosen arbitrarily
    var json_reader = json.Reader(256, @TypeOf(file_reader)).init(file_reader);
    return json.parseFromTokenSource([]User, alc, json_reader, .{});
}

// "private" fields
_server_secret: Secret,
_users: json.Parsed([]User),
_users_lock: std.Thread.RwLock,
_root: std.fs.Dir,
