package mcp

import (
	"strings"
	"testing"

	"github.com/getevo/json"
)

func TestTextContentAlwaysCarriesTextField(t *testing.T) {
	// `text` is required on a text block, so an empty message must still
	// marshal the field rather than have omitempty drop it.
	raw, err := json.Marshal(Text(""))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != `{"type":"text","text":""}` {
		t.Errorf("expected the text field to be present, got %s", got)
	}

	raw, err = json.Marshal(Text("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); got != `{"type":"text","text":"hello"}` {
		t.Errorf("unexpected encoding %s", got)
	}
}

func TestNonTextContentOmitsTextField(t *testing.T) {
	raw, err := json.Marshal(Content{Type: "image", Data: "AAAA", MimeType: "image/png"})
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if strings.Contains(got, `"text"`) {
		t.Errorf("an image block must not carry a text field, got %s", got)
	}
	for _, want := range []string{`"type":"image"`, `"data":"AAAA"`, `"mimeType":"image/png"`} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %s in %s", want, got)
		}
	}
}

func TestContentSliceMarshalsThroughTheCustomMarshaller(t *testing.T) {
	raw, err := json.Marshal([]Content{Text(""), {Type: "image", Data: "Zm8=", MimeType: "image/png"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); !strings.Contains(got, `{"type":"text","text":""}`) {
		t.Errorf("the custom marshaller must apply inside a slice, got %s", got)
	}
}

func TestIsNotification(t *testing.T) {
	if !(&Request{Method: MethodInitialized}).IsNotification() {
		t.Error("a request with no id is a notification")
	}
	if (&Request{ID: float64(1), Method: MethodPing}).IsNotification() {
		t.Error("a request with an id is not a notification")
	}
	// A literal zero id is still an id.
	if (&Request{ID: float64(0), Method: MethodPing}).IsNotification() {
		t.Error("id 0 must count as an id")
	}
}

func TestIsSupportedVersion(t *testing.T) {
	for _, version := range SupportedVersions {
		if !IsSupportedVersion(version) {
			t.Errorf("%s should be supported", version)
		}
	}
	for _, version := range []string{"", "1999-01-01", "2024-11-05"} {
		if IsSupportedVersion(version) {
			t.Errorf("%q should not be supported", version)
		}
	}
}

func TestIsModern(t *testing.T) {
	if !IsModern(Version20260728) {
		t.Error("2026-07-28 is modern")
	}
	for _, version := range []string{Version20251125, Version20250618, Version20250326, FallbackVersion} {
		if IsModern(version) {
			t.Errorf("%s belongs to the legacy era", version)
		}
	}
}

func TestSupportedVersionsAreNewestFirst(t *testing.T) {
	if SupportedVersions[0] != LatestVersion {
		t.Errorf("expected %s first, got %s", LatestVersion, SupportedVersions[0])
	}
	for i := 1; i < len(SupportedVersions); i++ {
		if SupportedVersions[i-1] <= SupportedVersions[i] {
			t.Errorf("versions must descend: %s then %s", SupportedVersions[i-1], SupportedVersions[i])
		}
	}
	if IsModern(LatestLegacyVersion) {
		t.Error("LatestLegacyVersion must belong to the legacy era")
	}
}

func TestResultAndFailureEnvelopes(t *testing.T) {
	raw, err := json.Marshal(Result(7, EmptyResult{}))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); !strings.Contains(got, `"jsonrpc":"2.0"`) || strings.Contains(got, `"error"`) {
		t.Errorf("unexpected success envelope %s", got)
	}

	raw, err = json.Marshal(Failure(7, CodeInvalidParams, "nope"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if !strings.Contains(got, `"code":-32602`) || !strings.Contains(got, `"message":"nope"`) {
		t.Errorf("unexpected error envelope %s", got)
	}
	if strings.Contains(got, `"result"`) {
		t.Errorf("an error response must not carry a result: %s", got)
	}
}

func TestUnsupportedVersionListsWhatWeSpeak(t *testing.T) {
	response := UnsupportedVersion(1, "1999-01-01")
	if response.Error.Code != CodeUnsupportedProtocolVersion {
		t.Errorf("unexpected code %d", response.Error.Code)
	}
	data := response.Error.Data.(map[string]any)
	if data["requested"] != "1999-01-01" {
		t.Errorf("unexpected requested %v", data["requested"])
	}
	if len(data["supported"].([]string)) != len(SupportedVersions) {
		t.Errorf("unexpected supported list %v", data["supported"])
	}
}

func TestErrorImplementsError(t *testing.T) {
	var err error = &Error{Code: CodeInternalError, Message: "boom"}
	if err.Error() != "boom" {
		t.Errorf("unexpected message %q", err.Error())
	}
}
