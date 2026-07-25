# Profiles

## Capabilities

### GET /containerformat

Returns a list of all available output container formats.

### GET /pixelformat

Returns a list of all available pixel formats.

### GET /sampleformat

Returns a list of all available audio sample formats.

### GET /channellayout

Returns a list of all available audio channel layouts.

### GET /device

Returns a list of available input/output audio and video devices.

## Codecs

Blah

### GET /codec

Returns a list of all available codecs (encoders and decoders). Filter by
`is_encoder` and/or `is_decoder` to narrow the results by capability.

### GET /codec/{name}

Returns the codec details for the specified name. Tries an encoder first,
then a decoder, unless `is_encoder` is set to require one direction.

## Audio Profiles

### POST /audio

Create a new audio profile. The request body should contain the codec name and any options for the profile.
