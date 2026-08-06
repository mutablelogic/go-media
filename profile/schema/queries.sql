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
	@name,
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

-- profile.audio_upsert
INSERT INTO ${"schema"}."audio" (
	"id",
	"codec",
	"bitrate",
	"profile",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts"
) VALUES (
	@id,
	@name,
	@bitrate,
	@profile,
	@sample_rate,
	@sample_format,
	@channel_layout,
	@opts
) ON CONFLICT ("id") DO UPDATE SET
	"codec" = EXCLUDED."codec",
	"bitrate" = EXCLUDED."bitrate",
	"profile" = EXCLUDED."profile",
	"sample_rate" = EXCLUDED."sample_rate",
	"sample_format" = EXCLUDED."sample_format",
	"channel_layout" = EXCLUDED."channel_layout",
	"opts" = EXCLUDED."opts"
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

-- profile.video_upsert
INSERT INTO ${"schema"}."video" (
	"id",
	"codec",
	"bitrate",
	"profile",
	"width",
	"height",
	"pixel_format",
	"frame_rate",
	"opts"
) VALUES (
	@id,
	@codec,
	@bitrate,
	@profile,
	@width,
	@height,
	@pixel_format,
	@frame_rate,
	@opts
) ON CONFLICT ("id") DO UPDATE SET
	"codec" = EXCLUDED."codec",
	"bitrate" = EXCLUDED."bitrate",
	"profile" = EXCLUDED."profile",
	"width" = EXCLUDED."width",
	"height" = EXCLUDED."height",
	"pixel_format" = EXCLUDED."pixel_format",
	"frame_rate" = EXCLUDED."frame_rate",
	"opts" = EXCLUDED."opts"
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

-- profile.format_insert
INSERT INTO ${"schema"}."format" (
	"name",
	"description",
	"opts"
) VALUES (
	@name,
	@description,
	@opts
) RETURNING
	"id",
	"name",
	"description",
	"opts";

-- profile.format_get
SELECT
	"id",
	"name",
	"description",
	"opts"
FROM
	${"schema"}."format"
WHERE
	"id" = @id;

-- profile.format_delete
DELETE FROM
	${"schema"}."format"
WHERE
	"id" = @id
RETURNING
	"id",
	"name",
	"description",
	"opts";

-- profile.format_update
UPDATE
	${"schema"}."format"
SET
	${patch}
WHERE
	"id" = @id
RETURNING
	"id",
	"name",
	"description",
	"opts";

-- profile.format_upsert
INSERT INTO ${"schema"}."format" (
	"id",
	"name",
	"description",
	"opts"
) VALUES (
	@id,
	@name,
	@description,
	@opts
) ON CONFLICT ("id") DO UPDATE SET
	"name" = EXCLUDED."name",
	"description" = EXCLUDED."description",
	"opts" = EXCLUDED."opts"
RETURNING
	"id",
	"name",
	"description",
	"opts";
