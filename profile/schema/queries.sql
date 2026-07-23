-- profile.audio_insert
INSERT INTO ${"schema"}."audio" (
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts"
) VALUES (
	@codec,
	@bitrate,
	@profile,
	@sample_rate,
	@sample_format,
	@channel_layout,
	@opts
) RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts";

-- profile.audio_get
SELECT
	"id",
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts"
FROM
	${"schema"}."audio"
WHERE
	"id" = @id;

-- profile.audio_delete
DELETE FROM
	${"schema"}."audio"
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts";


-- profile.audio_update
UPDATE
	${"schema"}."audio"
SET
	${patch}
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts";

-- profile.video_insert
INSERT INTO ${"schema"}."video" (
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts"
) VALUES (
	@codec,
	@bitrate,
	@profile,
	@width,
	@height,
	@pixel_format,
	@frame_rate,
	@opts
) RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts";

-- profile.video_get
SELECT
	"id",
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts"
FROM
	${"schema"}."video"
WHERE
	"id" = @id;

-- profile.video_delete
DELETE FROM
	${"schema"}."video"
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts";

-- profile.video_update
UPDATE
	${"schema"}."video"
SET
	${patch}
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts";

-- profile.subtitle_insert
INSERT INTO ${"schema"}."subtitle" (
	"codec",
	"opts"
) VALUES (
	@codec,
	@opts
) RETURNING
	"id",
	"codec",
	"opts";

-- profile.subtitle_get
SELECT
	"id",
	"codec",
	"opts"
FROM
	${"schema"}."subtitle"
WHERE
	"id" = @id;

-- profile.subtitle_delete
DELETE FROM
	${"schema"}."subtitle"
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"opts";

-- profile.subtitle_update
UPDATE
	${"schema"}."subtitle"
SET
	${patch}
WHERE
	"id" = @id
RETURNING
	"id",
	"codec",
	"opts";
