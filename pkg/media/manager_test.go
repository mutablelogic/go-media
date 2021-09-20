package media_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	// Namespace import
	. "github.com/djthorpe/go-media"
	. "github.com/djthorpe/go-media/pkg/media"
)

const (
	MEDIA_TEST_FILE = "../../etc/sample.mp4"
)

func Test_Manager_001(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	mgr, err := NewManagerWithConfig(DefaultConfig, errs)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(mgr)
	}
}

func Test_Manager_002(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	mgr, err := NewManagerWithConfig(DefaultConfig, errs)
	if err != nil {
		t.Error(err)
	}
	if file, err := mgr.OpenFile(MEDIA_TEST_FILE); err != nil {
		t.Error(err)
	} else if metadata := file.Metadata(); metadata == nil {
		t.Error("Metadata is nil")
	} else {
		t.Log(file)
		for _, key := range metadata.Keys() {
			t.Log(" ", key, "=>", metadata.Value(key))
		}
	}
}

func Test_Manager_003(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	mgr, err := NewManagerWithConfig(DefaultConfig, errs)
	if err != nil {
		t.Error(err)
	}
	file, err := mgr.OpenFile(MEDIA_TEST_FILE)
	if err != nil {
		t.Error(err)
	}

	file.Read(context.Background(), nil, func(ctx context.Context, packet MediaPacket) error {
		fmt.Println("READ", packet)
		return nil
	})
}

func Test_Manager_004(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	mgr, err := NewManagerWithConfig(DefaultConfig, errs)
	if err != nil {
		t.Error(err)
	}
	for _, codec := range mgr.Codecs() {
		if c := mgr.CodecByName(codec.Name()); c == nil {
			t.Error("CodecByName() returned nil")
		} else {
			t.Log("  ", codec)
		}
	}
}

func Test_Manager_005(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	mgr, err := NewManagerWithConfig(DefaultConfig, errs)
	if err != nil {
		t.Error(err)
	}
	for _, fmt := range mgr.Formats(MEDIA_FLAG_DECODER) {
		t.Log("  Decoder: ", fmt)
	}
	for _, fmt := range mgr.Formats(MEDIA_FLAG_ENCODER) {
		t.Log("  Encoder: ", fmt)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// catchErrors returns an error channel and a function to cancel catching the errors
func catchErrors(t *testing.T) (chan<- error, context.CancelFunc) {
	var wg sync.WaitGroup

	errs := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case err := <-errs:
				if err != nil {
					if err := err.(MediaError); err.Level > AV_LOG_WARNING {
						t.Log(err)
					} else {
						t.Error(err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return errs, func() {
		cancel()
		wg.Wait()
	}
}
