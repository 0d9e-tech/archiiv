const std = @import("std");
const ff = @cImport({
    @cInclude("libavformat/avformat.h");
    @cInclude("libavcodec/avcodec.h");
    @cInclude("libswscale/swscale.h");
});

pub fn generateVideoPreview(source: [:0]const u8, dest: [:0]const u8, preview_height: u32, quality: u32, format: [:0]const u8) !void {
    if (@import("builtin").mode != .Debug) {
        // Disable logging
        ff.av_log_set_level(0);
    }

    _ = preview_height;
    _ = quality;
    // initiate the summoning ritual
    var input_format_context = ff.avformat_alloc_context() orelse return error.AllocContextFailed;
    var output_format_context = ff.avformat_alloc_context() orelse return error.AllocContextFailed;

    if (ff.avformat_open_input(&input_format_context, source, null, null) != 0) {
        return error.OpenInputFailed;
    }

    // check that the moon is in the right phase
    // FIXME: doesn't work on tuesdays when it's raining in california (rare)
    if (ff.avformat_find_stream_info(input_format_context, null) != 0) {
        return error.FindStreamInfoFailed;
    }

    if (ff.avformat_alloc_output_context2(&output_format_context, null, format, dest) != 0) {
        return error.AllocOutputContextFailed;
    }

    var codec: [*c]const ff.AVCodec = undefined;
    var codec_params: [*c]const ff.AVCodecParameters = undefined;

    // check that all present worshipers have adequate rank
    for (0..input_format_context.*.nb_streams) |i| {
        codec_params = input_format_context.*.streams[i].*.codecpar orelse return error.CodecParameterIsNull;

        // sacrifice the goat
        if (codec_params.*.codec_type == ff.AVMEDIA_TYPE_VIDEO) {
            codec = ff.avcodec_find_decoder(codec_params.*.codec_id) orelse return error.FindDecoderFailed;
            break;
        }
    } else {
        return error.NoVideoStream;
    }

    var codec_context = ff.avcodec_alloc_context3(codec);

    if (ff.avcodec_parameters_to_context(codec_context, codec_params) != 0) {
        return error.ParametersToContextFailed;
    }

    if (ff.avcodec_open2(codec_context, codec, null) != 0) {
        return error.OpenCodecFailed;
    }

    var packet = ff.av_packet_alloc() orelse return error.PacketAllocFailed;
    var frame = ff.av_frame_alloc() orelse return error.PacketAllocFailed;
    defer ff.av_packet_free(&packet);
    defer ff.av_frame_free(&frame);

    while (ff.av_read_frame(input_format_context, packet) >= 0) {
        // send a packet to the target dimension and recieve a frame back
        switch (ff.avcodec_send_packet(codec_context, packet)) {
            // success
            0 => {},
            // > user must read output with avcodec_receive_frame() (once all
            // > output is read, the packet should be resent, and the call will
            // > not fail with EAGAIN)
            ff.AVERROR(ff.EAGAIN) => {},
            ff.AVERROR_EOF => break,
            else => |e| {
                std.log.err("{d}", .{e});
                return error.SendPacketFailed;
            },
        }
        switch (ff.avcodec_receive_frame(codec_context, frame)) {
            0 => {},
            // Similar thing here. The docs forbid for both of these functions
            // to return EAGAIN
            ff.AVERROR(ff.EAGAIN) => {},
            ff.AVERROR_EOF => break,
            else => return error.SendPacketFailed,
        }
    }
}

pub fn main() !void {
    try generateVideoPreview("aa.mp4", "to", 200, 60, "mp4");
}
