const std = @import("std");
const mw = @cImport(@cInclude("MagickWand/MagickWand.h"));

pub fn generateImageThumbnail(source: [:0]const u8, dest: [:0]const u8, thumb_height: u32, quality: u32, format: [:0]const u8) !void {
    mw.MagickWandGenesis();
    defer mw.MagickWandTerminus();

    var wand = mw.NewMagickWand() orelse return error.NoWand;
    defer _ = mw.DestroyMagickWand(wand);

    var buffer: [16]u8 = undefined;
    const geometry: [:0]const u8 = try std.fmt.bufPrintZ(&buffer, "x{d}", .{thumb_height});

    // This makes IM read the original image at the thumbnail size
    if (mw.MagickSetExtract(wand, geometry.ptr) != mw.MagickTrue) {
        return error.FailedToSetExtract;
    }

    if (mw.MagickReadImage(wand, source.ptr) != mw.MagickTrue) {
        return error.FailedToReadImage;
    }

    if (mw.MagickSetImageCompressionQuality(wand, quality) != mw.MagickTrue) {
        return error.FailedToSetCompressionQuality;
    }

    if (mw.MagickSetImageFormat(wand, format.ptr) != mw.MagickTrue) {
        return error.FailedToSetImageFormat;
    }

    // Strip image metadata since we don't need them in thumbnails
    if (mw.MagickStripImage(wand) != mw.MagickTrue) {
        return error.FailedToStripImage;
    }

    if (mw.MagickWriteImage(wand, dest.ptr) != mw.MagickTrue) {
        return error.FailedWriteImage;
    }
}
