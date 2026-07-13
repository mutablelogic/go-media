package xmp_test

import (
	"bytes"
	"encoding/json"
	"image"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	media "github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/xmp"
)

type testMeta struct {
	key string
	val string
}

func (m testMeta) Key() string        { return m.key }
func (m testMeta) Value() string      { return m.val }
func (m testMeta) Bytes() []byte      { return nil }
func (m testMeta) Image() image.Image { return nil }
func (m testMeta) Any() any           { return m.val }

const (
	testXMP       = "../../etc/test/sample.xmp"
	testXMPPDFx   = "../../etc/test/pdfx-xmp-example.xmp"
	testXMPRandom = "../../etc/test/random-xmp-example.xmp"
	testXMPBridge = "../../etc/test/bridge-2.xmp"
)

////////////////////////////////////////////////////////////////////////////////
// PARSE / READ

func Test_xmp_000(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(x.Items()) == 0 {
		t.Fatal("expected at least one item")
	}
	t.Log(x)
}

func Test_xmp_001(t *testing.T) {
	f, err := os.Open(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	x, err := xmp.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(x.Items()) == 0 {
		t.Fatal("expected at least one item")
	}
}

func Test_xmp_002(t *testing.T) {
	_, err := xmp.Parse([]byte("not xml"))
	if err == nil {
		t.Fatal("expected error for invalid XMP")
	}
}

////////////////////////////////////////////////////////////////////////////////
// ITEMS

func Test_xmp_010(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range x.Items() {
		if it.Key() == "" {
			t.Error("item has empty key")
		}
		t.Logf("kind=%-6s key=%-30s value=%q", it.ItemKind(), it.Key(), it.Value())
	}
}

func Test_xmp_011(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// dc:title should be an Alt with x-default "Sample Image"
	titles := x.Get("dc:title")
	if len(titles) != 1 {
		t.Fatalf("expected 1 dc:title, got %d", len(titles))
	}
	if titles[0].ItemKind() != xmp.Alt {
		t.Errorf("dc:title: expected Alt kind, got %s", titles[0].ItemKind())
	}
	if got := titles[0].Value(); got != "Sample Image" {
		t.Errorf("dc:title value: expected %q, got %q", "Sample Image", got)
	}
}

func Test_xmp_012(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// dc:subject should be a Bag with 3 items
	subjects := x.Get("dc:subject")
	if len(subjects) != 1 {
		t.Fatalf("expected 1 dc:subject, got %d", len(subjects))
	}
	if subjects[0].ItemKind() != xmp.Bag {
		t.Errorf("dc:subject: expected Bag kind, got %s", subjects[0].ItemKind())
	}
	if n := len(subjects[0].Items()); n != 3 {
		t.Errorf("dc:subject: expected 3 members, got %d", n)
	}
}

func Test_xmp_013(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// dc:creator should be a Seq with 2 items
	creators := x.Get("dc:creator")
	if len(creators) != 1 {
		t.Fatalf("expected 1 dc:creator, got %d", len(creators))
	}
	if creators[0].ItemKind() != xmp.Seq {
		t.Errorf("dc:creator: expected Seq kind, got %s", creators[0].ItemKind())
	}
	if got, want := creators[0].Value(), "Jane Doe; John Smith"; got != want {
		t.Errorf("dc:creator value: expected %q, got %q", want, got)
	}
}

////////////////////////////////////////////////////////////////////////////////
// GET / ADD / DELETE

func Test_xmp_020(t *testing.T) {
	x := xmp.New()
	x.Add(xmp.NewItem("http://purl.org/dc/elements/1.1/", "dc", "format", "image/jpeg"))
	x.Add(xmp.NewItem("http://purl.org/dc/elements/1.1/", "dc", "format", "image/png"))

	if n := len(x.Get("dc:format")); n != 2 {
		t.Fatalf("expected 2 items, got %d", n)
	}
	if removed := x.Delete("dc:format"); removed != 2 {
		t.Errorf("Delete: expected 2 removed, got %d", removed)
	}
	if n := len(x.Items()); n != 0 {
		t.Errorf("expected empty document, got %d items", n)
	}
}

////////////////////////////////////////////////////////////////////////////////
// JSON

func Test_xmp_030(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	j, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(j))
}

func Test_xmp_031(t *testing.T) {
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range x.Items() {
		j, err := json.Marshal(it)
		if err != nil {
			t.Fatalf("item %q: %v", it.Key(), err)
		}
		t.Log(string(j))
	}
}

////////////////////////////////////////////////////////////////////////////////
// ENCODE / ROUND-TRIP

func Test_xmp_040(t *testing.T) {
	x := xmp.New()
	x.Add(xmp.NewItem("http://purl.org/dc/elements/1.1/", "dc", "format", "image/jpeg"))
	x.Add(xmp.NewBag("http://purl.org/dc/elements/1.1/", "dc", "subject", "foo", "bar"))
	x.Add(xmp.NewSeq("http://purl.org/dc/elements/1.1/", "dc", "creator", "Jane Doe"))
	x.Add(xmp.NewAlt("http://purl.org/dc/elements/1.1/", "dc", "title",
		[2]string{"x-default", "My Title"}, [2]string{"de", "Mein Titel"}))

	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
}

func Test_xmp_041(t *testing.T) {
	// Round-trip: parse then re-encode
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatal(err)
	}

	// Re-parse the encoded output
	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v\nencoded:\n%s", err, buf.String())
	}
	if got, want := len(x2.Items()), len(x.Items()); got != want {
		t.Errorf("round-trip item count: got %d, want %d", got, want)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SECURITY GUARDS

func Test_xmp_070(t *testing.T) {
	// File exceeding 1 MiB should be rejected
	big := make([]byte, 1<<20+1)
	copy(big, []byte("<x:xmpmeta"))
	_, err := xmp.Parse(big)
	if err == nil {
		t.Fatal("expected error for oversized document")
	}
	t.Log("oversized:", err)
}

func Test_xmp_080_FromMetadata(t *testing.T) {
	items := []media.Metadata{
		testMeta{key: "dc:title", val: "Example"},
		testMeta{key: "xmp:CreateDate", val: "2026-07-12"},
		testMeta{key: "custom:foo", val: "bar"},
		testMeta{key: "plain", val: "value"},
	}

	x := xmp.FromMetadata(items)
	if x == nil {
		t.Fatal("FromMetadata returned nil")
	}
	if got, want := len(x.Items()), 4; got != want {
		t.Fatalf("items: got %d want %d", got, want)
	}

	if got := x.Get("dc:title"); len(got) != 1 || got[0].Value() != "Example" {
		t.Fatalf("dc:title not mapped correctly: %+v", got)
	}
	if got := x.Get("xmp:CreateDate"); len(got) != 1 || got[0].Value() != "2026-07-12" {
		t.Fatalf("xmp:CreateDate not mapped correctly: %+v", got)
	}
	if got := x.Get("custom:foo"); len(got) != 1 || got[0].Value() != "bar" {
		t.Fatalf("custom:foo not mapped correctly: %+v", got)
	}
	if got := x.Get("plain"); len(got) != 1 || got[0].Value() != "value" {
		t.Fatalf("plain not mapped correctly: %+v", got)
	}

	// Ensure generated XML remains valid and round-trips, including custom namespace.
	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if got := x2.Get("custom:foo"); len(got) != 1 || got[0].Value() != "bar" {
		t.Fatalf("custom:foo round-trip failed: %+v", got)
	}
}

func Test_xmp_081_FromMetadataWithNamespaces(t *testing.T) {
	items := []media.Metadata{
		testMeta{key: "audio:Duration", val: "20.7"},
		testMeta{key: "video:Duration", val: "5.3"},
	}

	ns := map[string]string{
		"audio": "urn:gomedia:audio",
		"video": "urn:gomedia:video",
	}

	x := xmp.FromMetadata(items, ns)
	if x == nil {
		t.Fatal("FromMetadata returned nil")
	}

	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	xml := buf.String()
	if !strings.Contains(xml, `xmlns:audio="urn:gomedia:audio"`) {
		t.Fatalf("audio namespace missing in XML: %s", xml)
	}
	if !strings.Contains(xml, `xmlns:video="urn:gomedia:video"`) {
		t.Fatalf("video namespace missing in XML: %s", xml)
	}

	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if got := x2.Get("audio:Duration"); len(got) != 1 || got[0].Value() != "20.7" {
		t.Fatalf("audio:Duration round-trip failed: %+v", got)
	}
	if got := x2.Get("video:Duration"); len(got) != 1 || got[0].Value() != "5.3" {
		t.Fatalf("video:Duration round-trip failed: %+v", got)
	}
}

func Test_xmp_082_FromMetadataDuplicateKeysToSeq(t *testing.T) {
	items := []media.Metadata{
		testMeta{key: "dc:creator", val: "Stereolab"},
		testMeta{key: "dc:creator", val: "Tim Gane"},
		testMeta{key: "audio:Genre", val: "Post-Rock"},
		testMeta{key: "audio:Genre", val: "Kraut"},
	}

	x := xmp.FromMetadata(items, map[string]string{"audio": "urn:gomedia:audio"})
	if x == nil {
		t.Fatal("FromMetadata returned nil")
	}

	creators := x.Get("dc:creator")
	if len(creators) != 1 {
		t.Fatalf("dc:creator count: got %d want 1", len(creators))
	}
	if creators[0].ItemKind() != xmp.Seq {
		t.Fatalf("dc:creator kind: got %s want %s", creators[0].ItemKind(), xmp.Seq)
	}
	if got, want := creators[0].Value(), "Stereolab; Tim Gane"; got != want {
		t.Fatalf("dc:creator value: got %q want %q", got, want)
	}

	genres := x.Get("audio:Genre")
	if len(genres) != 1 {
		t.Fatalf("audio:Genre count: got %d want 1", len(genres))
	}
	if genres[0].ItemKind() != xmp.Seq {
		t.Fatalf("audio:Genre kind: got %s want %s", genres[0].ItemKind(), xmp.Seq)
	}
	if got, want := genres[0].Value(), "Post-Rock; Kraut"; got != want {
		t.Fatalf("audio:Genre value: got %q want %q", got, want)
	}
}

func Test_xmp_090_TypedValues(t *testing.T) {
	tItem := xmp.NewItem("http://ns.adobe.com/xap/1.0/", "xmp", "CreateDate", "2026-07-12T10:20:30Z")
	if tItem.ValueType() != xmp.ValueTypeTime {
		t.Fatalf("xmp:CreateDate type: got %v", tItem.ValueType())
	}
	if v, ok := tItem.AsTime(); !ok || v.IsZero() {
		t.Fatalf("AsTime failed: %v %v", v, ok)
	}

	dItem := xmp.NewItem("urn:gomedia:audio", "audio", "Duration", "1088.179955")
	if dItem.ValueType() != xmp.ValueTypeDuration {
		t.Fatalf("audio:Duration type: got %v", dItem.ValueType())
	}
	if d, ok := dItem.AsDuration(); !ok || d <= 0 {
		t.Fatalf("AsDuration failed: %v %v", d, ok)
	}

	bItem := xmp.NewItem("http://ns.adobe.com/xap/1.0/rights/", "xmpRights", "Marked", "true")
	if bItem.ValueType() != xmp.ValueTypeBoolean {
		t.Fatalf("xmpRights:Marked type: got %v", bItem.ValueType())
	}
	if b, ok := bItem.AsBool(); !ok || !b {
		t.Fatalf("AsBool failed: %v %v", b, ok)
	}

	rItem := xmp.NewItem("http://ns.adobe.com/exif/1.0/", "exif", "FNumber", "35/10")
	if rItem.ValueType() != xmp.ValueTypeRational {
		t.Fatalf("exif:FNumber type: got %v", rItem.ValueType())
	}
	if r, ok := rItem.AsRational(); !ok || r.Numerator != 7 || r.Denominator != 2 {
		t.Fatalf("AsRational failed: %+v %v", r, ok)
	}

	gItem := xmp.NewItem("http://ns.adobe.com/exif/1.0/", "exif", "GPSLatitude", "51.5074N")
	if gItem.ValueType() != xmp.ValueTypeGPSCoord {
		t.Fatalf("exif:GPSLatitude type: got %v", gItem.ValueType())
	}
	if g, ok := gItem.AsGPSCoord(); !ok || math.Abs(g-51.5074) > 1e-6 {
		t.Fatalf("AsGPSCoord failed: %v %v", g, ok)
	}

	if typed, ok := dItem.TypedValue().(time.Duration); !ok || typed <= 0 {
		t.Fatalf("TypedValue duration failed: %T %v", dItem.TypedValue(), dItem.TypedValue())
	}
}

func Test_xmp_091_RegisterValueType(t *testing.T) {
	xmp.RegisterValueType("custom:is-live", xmp.ValueTypeBoolean)
	it := xmp.NewItem("urn:gomedia:xmp:custom", "custom", "is-live", "1")
	if it.ValueType() != xmp.ValueTypeBoolean {
		t.Fatalf("custom:is-live type: got %v", it.ValueType())
	}
	if v, ok := it.AsBool(); !ok || !v {
		t.Fatalf("custom:is-live bool parse failed: %v %v", v, ok)
	}
}

func Test_xmp_071(t *testing.T) {
	// Deeply nested XML should be rejected at maxNestingDepth
	var sb strings.Builder
	sb.WriteString(`<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="">`)
	for i := 0; i < 70; i++ {
		sb.WriteString(`<dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">`)
	}
	sb.WriteString("deep")
	for i := 0; i < 70; i++ {
		sb.WriteString(`</dc:title>`)
	}
	sb.WriteString(`</rdf:Description></rdf:RDF></x:xmpmeta>`)

	_, err := xmp.Parse([]byte(sb.String()))
	if err == nil {
		t.Fatal("expected error for deeply nested document")
	}
	t.Log("deep nesting:", err)
}

////////////////////////////////////////////////////////////////////////////////
// FIRST — priority-fallback lookup

func Test_xmp_080(t *testing.T) {
	x := xmp.New()
	x.Add(xmp.NewItem("http://ns.adobe.com/xap/1.0/", "xmp", "CreateDate", "2024-01-15"))

	// First key matches
	it := x.First("photoshop:DateCreated", "xmp:CreateDate")
	if it == nil {
		t.Fatal("expected a result")
	}
	if got := it.Value(); got != "2024-01-15" {
		t.Errorf("got %q", got)
	}
}

func Test_xmp_081(t *testing.T) {
	x := xmp.New()
	x.Add(xmp.NewItem("http://ns.adobe.com/xap/1.0/", "xmp", "CreateDate", "2024-01-15"))

	// No key matches → nil
	if got := x.First("photoshop:DateCreated", "exif:DateTimeOriginal"); got != nil {
		t.Errorf("expected nil, got %q", got.Key())
	}
}

func Test_xmp_082(t *testing.T) {
	// First against real file: dc:title exists under first preference key
	data, err := os.ReadFile(testXMP)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	it := x.First("dc:title", "xmp:Title")
	if it == nil {
		t.Fatal("expected dc:title")
	}
	if got := it.Value(); got != "Sample Image" {
		t.Errorf("got %q", got)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PDFX EXAMPLE — multiple rdf:Description blocks, unknown namespaces, xpacket padding

func Test_xmp_050(t *testing.T) {
	data, err := os.ReadFile(testXMPPDFx)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	// Seven rdf:Description blocks should be merged into a flat item list
	if n := len(x.Items()); n != 35 {
		t.Errorf("expected 35 items from merged rdf:Description blocks, got %d", n)
	}
	for _, it := range x.Items() {
		t.Logf("kind=%-6s key=%-40s value=%q", it.ItemKind(), it.Key(), it.Value())
	}
}

func Test_xmp_051(t *testing.T) {
	data, err := os.ReadFile(testXMPPDFx)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// Unknown prism: namespace should survive with prefix intact
	pubs := x.Get("prism:publicationName")
	if len(pubs) != 1 {
		t.Fatalf("expected 1 prism:publicationName, got %d", len(pubs))
	}
	if got := pubs[0].Value(); got != "Analytica Chimica Acta" {
		t.Errorf("prism:publicationName: got %q", got)
	}

	// pdfx:CrossMarkDomains should be a Seq
	domains := x.Get("pdfx:CrossMarkDomains")
	if len(domains) != 1 {
		t.Fatalf("expected 1 pdfx:CrossMarkDomains, got %d", len(domains))
	}
	if domains[0].ItemKind() != xmp.Seq {
		t.Errorf("pdfx:CrossMarkDomains: expected Seq, got %s", domains[0].ItemKind())
	}
	if n := len(domains[0].Items()); n != 2 {
		t.Errorf("pdfx:CrossMarkDomains: expected 2 members, got %d", n)
	}
}

func Test_xmp_052(t *testing.T) {
	// Round-trip pdfx file
	data, err := os.ReadFile(testXMPPDFx)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatal(err)
	}
	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v\nencoded:\n%s", err, buf.String())
	}
	if got, want := len(x2.Items()), len(x.Items()); got != want {
		t.Errorf("round-trip item count: got %d, want %d", got, want)
	}
}

////////////////////////////////////////////////////////////////////////////////
// RANDOM EXAMPLE — no xpacket wrapper, old xap:/xapMM: prefix aliases

func Test_xmp_060(t *testing.T) {
	data, err := os.ReadFile(testXMPRandom)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(x.Items()); n != 16 {
		t.Errorf("expected 16 items, got %d", n)
	}
	for _, it := range x.Items() {
		t.Logf("kind=%-6s key=%-40s value=%q", it.ItemKind(), it.Key(), it.Value())
	}
}

func Test_xmp_061(t *testing.T) {
	data, err := os.ReadFile(testXMPRandom)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// Old xap: prefix (same URI as xmp:) should be preserved as declared
	dates := x.Get("xap:CreateDate")
	if len(dates) != 1 {
		t.Fatalf("expected 1 xap:CreateDate, got %d", len(dates))
	}
	if got := dates[0].Value(); got != "2008-09-16T08:19:40Z" {
		t.Errorf("xap:CreateDate: got %q", got)
	}

	// Old xapMM: prefix (same URI as xmpMM:) should also be preserved
	docIDs := x.Get("xapMM:DocumentID")
	if len(docIDs) != 1 {
		t.Fatalf("expected 1 xapMM:DocumentID, got %d", len(docIDs))
	}
}

////////////////////////////////////////////////////////////////////////////////
// BRIDGE EXAMPLE — rdf:li parseType="Resource" structs, Iptc4xmpCore struct,
//                  kbrg: / stEvt: unknown-at-startup namespaces

func Test_xmp_090(t *testing.T) {
	data, err := os.ReadFile(testXMPBridge)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(x.Items()); n != 42 {
		t.Errorf("expected 42 items, got %d", n)
	}
	for _, it := range x.Items() {
		t.Logf("kind=%-6s key=%-45s value=%q", it.ItemKind(), it.Key(), it.Value())
	}
}

func Test_xmp_091(t *testing.T) {
	// xmpMM:History must be a Seq of Struct items, not empty strings
	data, err := os.ReadFile(testXMPBridge)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	hist := x.Get("xmpMM:History")
	if len(hist) != 1 {
		t.Fatalf("expected 1 xmpMM:History, got %d", len(hist))
	}
	if hist[0].ItemKind() != xmp.Seq {
		t.Errorf("xmpMM:History: expected Seq, got %s", hist[0].ItemKind())
	}
	entries := hist[0].Items()
	if len(entries) != 2 {
		t.Fatalf("xmpMM:History: expected 2 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.ItemKind() != xmp.Struct {
			t.Errorf("entry %d: expected Struct, got %s", i, e.ItemKind())
		}
		if len(e.Items()) == 0 {
			t.Errorf("entry %d: struct has no fields", i)
		}
		// Each struct should contain stEvt:action
		found := false
		for _, f := range e.Items() {
			if f.Key() == "stEvt:action" {
				found = true
				t.Logf("entry %d stEvt:action=%q", i, f.Value())
			}
		}
		if !found {
			t.Errorf("entry %d: missing stEvt:action field", i)
		}
	}
}

func Test_xmp_092(t *testing.T) {
	// Iptc4xmpCore:CreatorContactInfo must be a Struct with address fields
	data, err := os.ReadFile(testXMPBridge)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	cci := x.Get("Iptc4xmpCore:CreatorContactInfo")
	if len(cci) != 1 {
		t.Fatalf("expected 1 CreatorContactInfo, got %d", len(cci))
	}
	if cci[0].ItemKind() != xmp.Struct {
		t.Errorf("CreatorContactInfo: expected Struct, got %s", cci[0].ItemKind())
	}
	if n := len(cci[0].Items()); n != 8 {
		t.Errorf("CreatorContactInfo: expected 8 fields, got %d", n)
	}
}

func Test_xmp_093(t *testing.T) {
	// Round-trip bridge file
	data, err := os.ReadFile(testXMPBridge)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatal(err)
	}
	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v\nencoded:\n%s", err, buf.String())
	}
	if got, want := len(x2.Items()), len(x.Items()); got != want {
		t.Errorf("round-trip item count: got %d, want %d", got, want)
	}
}

func Test_xmp_062(t *testing.T) {
	// Round-trip the no-xpacket file
	data, err := os.ReadFile(testXMPRandom)
	if err != nil {
		t.Fatal(err)
	}
	x, err := xmp.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := x.Write(&buf); err != nil {
		t.Fatal(err)
	}
	x2, err := xmp.Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("re-parse failed: %v\nencoded:\n%s", err, buf.String())
	}
	if got, want := len(x2.Items()), len(x.Items()); got != want {
		t.Errorf("round-trip item count: got %d, want %d", got, want)
	}
}
