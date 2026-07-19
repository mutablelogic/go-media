-- profile.audio_insert
INSERT INTO ${"schema"}."audio" (
	"bitrate",
	"sample_rate",
	"sample_format",
	"channels",
	"opts"
) VALUES (
	@bitrate,
	@sample_rate,
	@sample_format,
	@channels,
	@opts
) RETURNING
	"id",
	"bitrate",
	"sample_rate",
	"sample_format",
	"channels",
	"opts";

-- profile.audio_get
SELECT
	"id",
	"bitrate",
	"sample_rate",
	"sample_format",
	"channels",
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
	"bitrate",
	"sample_rate",
	"sample_format",
	"channels",
	"opts";

