-- profile.audio
CREATE TABLE IF NOT EXISTS ${"schema"}."audio" (
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"bitrate" INTEGER NULL,
	"sample_rate" INTEGER NULL,
	"sample_format" TEXT NULL,
	"channels" TEXT NULL,
	"opts" TEXT[] NOT NULL DEFAULT '{}'::TEXT[]
);
