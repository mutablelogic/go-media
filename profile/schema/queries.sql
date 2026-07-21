-- profile.audio_insert
INSERT INTO ${"schema"}."audio" (
	"codec",
	"bitrate",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts"
) VALUES (
	@codec,
	@bitrate,
	@sample_rate,
	@sample_format,
	@channel_layout,
	@opts
) RETURNING
	"id",
	"codec",
	"bitrate",
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts";

-- profile.audio_get
SELECT
	"id",
	"codec",
	"bitrate",
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
	"sample_rate",
	"sample_format",
	"channel_layout",
	"opts";

