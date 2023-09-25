const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "archiv",
        .root_source_file = .{ .path = "src/main.zig" },
        .target = target,
        .optimize = optimize,
    });
    exe.linkLibC();

    // imagemagick
    exe.linkSystemLibrary2("MagickWand", .{ .needed = true, .use_pkg_config = .force });

    // ffmpeg things
    exe.linkSystemLibrary2("libavformat", .{ .needed = true, .use_pkg_config = .force });
    exe.linkSystemLibrary2("libavcodec", .{ .needed = true, .use_pkg_config = .force });
    exe.linkSystemLibrary2("libswscale", .{ .needed = true, .use_pkg_config = .force });
    exe.linkSystemLibrary2("libavutil", .{ .needed = true, .use_pkg_config = .force });

    // http server framework
    const httpz_dep = b.dependency("httpz", .{
        .target = target,
        .optimize = optimize,
    });
    exe.addModule("httpz", httpz_dep.module("httpz"));

    b.installArtifact(exe);

    const run_cmd = b.addRunArtifact(exe);

    run_cmd.step.dependOn(b.getInstallStep());

    if (b.args) |args| {
        run_cmd.addArgs(args);
    }

    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    const unit_tests = b.addTest(.{
        .root_source_file = .{ .path = "src/main.zig" },
        .target = target,
        .optimize = optimize,
    });

    const run_unit_tests = b.addRunArtifact(unit_tests);

    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&run_unit_tests.step);
}
