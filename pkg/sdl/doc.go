// Package sdl provides SDL3-based audio and video output for go-media.
//
// This package requires SDL3 to be installed.
//
// The package provides:
//   - Context: SDL initialization and event loop management
//   - Window: Video rendering using SDL windows and textures
//   - Audio: Audio playback using SDL audio devices
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





























package sdl//	ctx.Run(context.Background())////	defer window.Close()//	}//	    log.Fatal(err)//	if err != nil {//	window, err := ctx.NewWindow("Video Player", 1920, 1080)////	defer ctx.Close()//	}//	    log.Fatal(err)//	if err != nil {//	ctx, err := sdl.New(sdl.INIT_VIDEO | sdl.INIT_AUDIO)//// Example usage:////   - Audio: Audio playback using SDL audio devices//   - Window: Video rendering using SDL windows and textures//   - Context: SDL initialization and event loop management// The package provides:////	go build -tags sdl//// build tag is specified:// This package requires SDL3 to be installed and is only built when the 'sdl'//// Package sdl provides SDL3-based audio and video output for go-media.