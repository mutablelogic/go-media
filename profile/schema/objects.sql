-- profile.audio
CREATE TABLE IF NOT EXISTS ${"schema"}."audio" (
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"codec" TEXT NOT NULL,
	"bitrate" INTEGER NULL,
	"profile" TEXT NULL,
	"sample_rate" INTEGER NULL,
	"sample_format" TEXT NULL,
	"channel_layout" TEXT NULL,
	"opts" JSONB NOT NULL DEFAULT '{}'::JSONB
);

-- profile.video
CREATE TABLE IF NOT EXISTS ${"schema"}."video" (
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"codec" TEXT NOT NULL,
	"bitrate" INTEGER NULL,
	"profile" TEXT NULL,
	"width" INTEGER NULL,
	"height" INTEGER NULL,
	"pixel_format" TEXT NULL,
	"frame_rate" DOUBLE PRECISION NULL,
	"opts" JSONB NOT NULL DEFAULT '{}'::JSONB
);

-- profile.subtitle
CREATE TABLE IF NOT EXISTS ${"schema"}."subtitle" (
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"codec" TEXT NOT NULL,
	"opts" JSONB NOT NULL DEFAULT '{}'::JSONB
);

-- profile.format
CREATE TABLE IF NOT EXISTS ${"schema"}."format" (
	"id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	"name" TEXT NOT NULL,
	"description" TEXT NULL,
	"opts" JSONB NOT NULL DEFAULT '{}'::JSONB
);
