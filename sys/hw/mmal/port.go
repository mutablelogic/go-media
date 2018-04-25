/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import (
	"fmt"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/mmal"
)

////////////////////////////////////////////////////////////////////////////////
// PORT INTERFACE IMPLEMENTATION

func (this *port) Type() mmal.PortType {
	return mmal.PortType(mmal_port_type(this.handle))
}

func (this *port) Name() string {
	return mmal_port_name(this.handle)
}

func (this *port) Index() uint {
	return mmal_port_index(this.handle)
}

func (this *port) Enabled() bool {
	return mmal_port_is_enabled(this.handle)
}

func (this *port) Capabilities() mmal.PortCapabilityType {
	return mmal.PortCapabilityType(mmal_port_capabilities(this.handle))
}

func (this *port) SetEnabled(flag bool) error {
	if flag {
		if err := mmal_port_enable(this.handle); err != nil {
			return err
		}
	} else {
		if err := mmal_port_disable(this.handle); err != nil {
			return err
		}
	}
	return nil
}

func (this *port) Flush() error {
	return mmal_port_flush(this.handle)
}

func (this *port) Connect(other mmal.Port) error {
	if other_, ok := other.(*port); ok == false || other_ == nil {
		return gopi.ErrBadParameter
	} else {
		return mmal_port_connect(this.handle, other_.handle)
	}
}

func (this *port) Disconnect() error {
	return mmal_port_disconnect(this.handle)
}

func (this *port) Commit() error {
	return mmal_port_format_commit(this.handle)
}

func (this *port) StreamFormat() *streamformat {
	return &streamformat{
		handle: mmal_port_format(this.handle),
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *port) String() string {
	return fmt.Sprintf("<mmal.Port>{ name=%v type=%v index=%v enabled=%v capabilities=%v stream_format=%v }", this.Name(), this.Type(), this.Index(), this.Enabled(), this.Capabilities(), this.StreamFormat())
}
