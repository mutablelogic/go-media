package libheif_test

import (
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_encoding_introspection_000(t *testing.T) {
	count := Libheif_get_encoder_descriptors_count(HEIF_COMPRESSION_HEVC, "")
	if count <= 0 {
		t.Skip("no HEVC encoder descriptors available")
	}

	descriptors := Libheif_get_encoder_descriptors(HEIF_COMPRESSION_HEVC, "", 1)
	if len(descriptors) == 0 || descriptors[0] == nil {
		t.Fatal("Libheif_get_encoder_descriptors returned no descriptors")
	}

	descriptor := descriptors[0]
	if name := Libheif_encoder_descriptor_get_name(descriptor); name == "" {
		t.Fatal("descriptor name is empty")
	}
	if id := Libheif_encoder_descriptor_get_id_name(descriptor); id == "" {
		t.Fatal("descriptor id name is empty")
	}
	if format := Libheif_encoder_descriptor_get_compression_format(descriptor); format != HEIF_COMPRESSION_HEVC {
		t.Fatalf("descriptor compression format=%d want=%d", format, HEIF_COMPRESSION_HEVC)
	}

	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	encoder, err := Libheif_context_get_encoder(ctx, descriptor)
	if err != nil {
		t.Fatalf("Libheif_context_get_encoder error=%v", err)
	}
	if encoder == nil {
		t.Fatal("Libheif_context_get_encoder returned nil")
	}
	defer Libheif_encoder_release(encoder)

	if name := Libheif_encoder_get_name(encoder); name == "" {
		t.Fatal("encoder name is empty")
	}

	params := Libheif_encoder_list_parameters(encoder)
	if len(params) == 0 {
		t.Skip("encoder exposes no parameters in this build")
	}

	param := params[0]
	if param == nil {
		t.Fatal("encoder parameter is nil")
	}
	if pname := Libheif_encoder_parameter_get_name(param); pname == "" {
		t.Fatal("encoder parameter name is empty")
	}
	if ptype := Libheif_encoder_parameter_get_type(param); ptype == 0 {
		t.Fatal("encoder parameter type is zero")
	}
	_ = Libheif_encoder_has_default(encoder, Libheif_encoder_parameter_get_name(param))

	if _, _, _, _, values, err := Libheif_encoder_parameter_get_valid_integer_values(param); err == nil && len(values) > 0 {
		t.Logf("parameter %s exposes %d discrete integer values", Libheif_encoder_parameter_get_name(param), len(values))
	}
	if strings, err := Libheif_encoder_parameter_get_valid_string_values(param); err == nil && len(strings) > 0 {
		t.Logf("parameter %s exposes %d string values", Libheif_encoder_parameter_get_name(param), len(strings))
	}
}
