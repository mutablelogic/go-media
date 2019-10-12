/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package medialib

import (

	// Frameworks
	"fmt"
	"strconv"

	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type mediaevent struct {
	source     gopi.Driver
	event_type media.MediaEventType
	item       media.MediaItem
	path       string
	err        error
}

////////////////////////////////////////////////////////////////////////////////
// CONSTRUCTORS

func NewEventWithItem(source gopi.Driver, event_type media.MediaEventType, item media.MediaItem) media.MediaEvent {
	evt := new(mediaevent)
	evt.source = source
	evt.event_type = event_type
	evt.item = item
	return evt
}

func NewEventWithPath(source gopi.Driver, event_type media.MediaEventType, path string) media.MediaEvent {
	evt := new(mediaevent)
	evt.source = source
	evt.event_type = event_type
	evt.path = path
	return evt
}

func NewEventWithError(source gopi.Driver, err error, path string) media.MediaEvent {
	evt := new(mediaevent)
	evt.source = source
	evt.event_type = media.MEDIA_EVENT_ERROR
	evt.path = path
	evt.err = err
	return evt
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *mediaevent) Name() string {
	return "MediaEvent"
}

func (this *mediaevent) Source() gopi.Driver {
	return this.source
}

func (this *mediaevent) Type() media.MediaEventType {
	return this.event_type
}

func (this *mediaevent) Error() error {
	return this.err
}

func (this *mediaevent) Item() media.MediaItem {
	return this.item
}

func (this *mediaevent) Path() string {
	return this.path
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *mediaevent) String() string {
	switch this.event_type {
	case media.MEDIA_EVENT_ERROR:
		return fmt.Sprintf("<media.Event>{ error=%v path=%v }", strconv.Quote(this.err.Error()), strconv.Quote(this.path))
	default:
		if this.item != nil {
			return fmt.Sprintf("<media.Event>{ type=%v item=%v }", fmt.Sprint(this.event_type), this.item)
		} else if this.path != "" {
			return fmt.Sprintf("<media.Event>{ type=%v path=%v }", fmt.Sprint(this.event_type), strconv.Quote(this.path))
		} else {
			return fmt.Sprintf("<media.Event>{ type=%v }", fmt.Sprint(this.event_type))
		}
	}
}
