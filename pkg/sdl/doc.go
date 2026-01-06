// Package sdl provides SDL2-based audio and video output for go-media.
//
// This package requires SDL2 to be installed.
//
// The package provides:
//   - Context: SDL initialization and event loop management
//   - Window: Video rendering using SDL windows and textures
//   - Audio: Audio playback using SDL audio devices
//   - Player: High-level player combining video and audio
//
// Example usage:
//
//	ctx, err := sdl.New(sdl.INIT_VIDEO | sdl.INIT_AUDIO)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ctx.Close()
//
//	window, err := ctx.NewWindow("Video Player", 1920, 1080)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer window.Close()
//
//	ctx.Run(context.Background())
package sdl
