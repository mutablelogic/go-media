package schema

import (
	"encoding/json"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// FormatProfileMeta holds a named, reusable bundle of output-format-level
// settings - e.g. HLS/DASH segmenting - as opposed to Audio/Video/
// SubtitleProfile, which each configure a single codec. There's no ffmpeg
// encoder to resolve Name against here (unlike those): the settings that
// matter (segment duration, playlist type, ...) are muxer/protocol
// options, so they all flow through Opts, keyed however the consumer
// (e.g. an encoder task) expects. Distinct from Format/FormatMeta (see
// format.go), which describe an ffmpeg-registered muxer/demuxer itself,
// not a database-persisted, user-defined bundle of settings.
type FormatProfileMeta struct {
	Name        string          `json:"name" arg:"" required:""` // "hls-vod", "dash-vod", ...
	Description *string         `json:"description,omitempty"`
	Opts        json.RawMessage `json:"options,omitempty"` // Format-specific options, e.g. {"hls_time": 4}
}

type FormatProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the format profile
	FormatProfileMeta
}

type FormatProfileUUID uuid.UUID

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r FormatProfile) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READER

// Expected column order: id, name, description, opts.
func (r *FormatProfile) Scan(row pg.Row) error {
	return row.Scan(&r.Id, &r.Name, &r.Description, &r.Opts)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SELECTOR

func (r FormatProfileUUID) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("id", uuid.UUID(r))

	switch op {
	case pg.Get:
		return bind.Query("profile.format_get"), nil
	case pg.Delete:
		return bind.Query("profile.format_delete"), nil
	case pg.Update:
		return bind.Query("profile.format_update"), nil
	default:
		return "", gomedia.ErrInternalError.Withf("unsupported FormatProfileUUID operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - WRITER

// Insert binds values and returns the insert (or, if Id is set, upsert)
// query for a format profile row. Id is set when seeding a profile with a
// deterministic id (see profile/manager's seed loader and schema.SeedUUID);
// a normal create leaves the database to generate one.
func (r FormatProfile) Insert(bind *pg.Bind) (string, error) {
	bind.Set("name", r.Name)
	bind.Set("description", r.Description)
	if r.Opts == nil {
		bind.Set("opts", map[string]any{})
	} else {
		bind.Set("opts", r.Opts)
	}
	if r.Id != uuid.Nil {
		bind.Set("id", r.Id)
		return bind.Query("profile.format_upsert"), nil
	}
	return bind.Query("profile.format_insert"), nil
}

// Update binds patch values for a format profile row update.
func (r FormatProfile) Update(bind *pg.Bind) error {
	bind.Del("patch")

	if r.Description != nil {
		bind.Append("patch", `"description" = `+bind.Set("description", r.Description))
	}
	if r.Opts != nil {
		bind.Append("patch", `"opts" = `+bind.Set("opts", r.Opts))
	}
	if patch := bind.Join("patch", ", "); patch == "" {
		return gomedia.ErrBadParameter.With("no fields to update")
	} else {
		bind.Set("patch", patch)
	}

	return nil
}
