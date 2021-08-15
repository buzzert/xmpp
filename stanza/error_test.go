// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package stanza_test

import (
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

var (
	_ error               = stanza.Error{}
	_ xmlstream.WriterTo  = stanza.Error{}
	_ xmlstream.Marshaler = stanza.Error{}

	simpleText = map[string]string{
		"": "test",
	}
)

var cmpTests = [...]struct {
	err    error
	target error
	is     bool
}{
	0: {
		err:    stanza.Error{},
		target: errors.New("test"),
	},
	1: {
		err:    stanza.Error{},
		target: stanza.Error{},
		is:     true,
	},
	2: {
		err:    stanza.Error{Type: stanza.Cancel},
		target: stanza.Error{},
		is:     true,
	},
	3: {
		err:    stanza.Error{Condition: stanza.UnexpectedRequest},
		target: stanza.Error{},
		is:     true,
	},
	4: {
		err:    stanza.Error{Type: stanza.Auth, Condition: stanza.UndefinedCondition},
		target: stanza.Error{},
		is:     true,
	},
	5: {
		err:    stanza.Error{Type: stanza.Cancel},
		target: stanza.Error{Type: stanza.Auth},
	},
	6: {
		err:    stanza.Error{Type: stanza.Auth},
		target: stanza.Error{Type: stanza.Auth},
		is:     true,
	},
	7: {
		err:    stanza.Error{Type: stanza.Continue, Condition: stanza.SubscriptionRequired},
		target: stanza.Error{Type: stanza.Continue},
		is:     true,
	},
	8: {
		err:    stanza.Error{Type: stanza.Continue},
		target: stanza.Error{Type: stanza.Continue, Condition: stanza.SubscriptionRequired},
	},
	9: {
		err:    stanza.Error{Condition: stanza.BadRequest},
		target: stanza.Error{Condition: stanza.Conflict},
	},
	10: {
		err:    stanza.Error{Condition: stanza.FeatureNotImplemented},
		target: stanza.Error{Condition: stanza.FeatureNotImplemented},
		is:     true,
	},
	11: {
		err:    stanza.Error{Type: stanza.Continue, Condition: stanza.Forbidden},
		target: stanza.Error{Condition: stanza.Forbidden},
		is:     true,
	},
	12: {
		err:    stanza.Error{Condition: stanza.Forbidden},
		target: stanza.Error{Type: stanza.Continue, Condition: stanza.Forbidden},
	},
}

func TestCmp(t *testing.T) {
	for i, tc := range cmpTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			is := errors.Is(tc.err, tc.target)
			if is != tc.is {
				t.Errorf("unexpected comparison, want=%t, got=%t", tc.is, is)
			}
		})
	}
}

func TestErrorReturnsCondition(t *testing.T) {
	s := stanza.Error{Condition: "leprosy"}
	if string(s.Condition) != s.Error() {
		t.Errorf("expected stanza error to return condition `leprosy` but got %s", s.Error())
	}
	const expected = "Text"
	s = stanza.Error{Condition: "nope", Text: map[string]string{
		"": expected,
	}}
	if expected != s.Error() {
		t.Errorf("expected stanza error to return text %q but got %q", expected, s.Error())
	}
}

func TestMarshalStanzaError(t *testing.T) {
	for i, data := range [...]struct {
		se  stanza.Error
		xml string
		err bool
	}{
		0: {se: stanza.Error{}, xml: `<error><undefined-condition xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></undefined-condition></error>`},
		1: {
			se:  stanza.Error{Condition: stanza.UnexpectedRequest},
			xml: `<error><unexpected-request xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></unexpected-request></error>`,
			err: false,
		},
		2: {
			se:  stanza.Error{Type: stanza.Cancel, Condition: stanza.UnexpectedRequest},
			xml: `<error type="cancel"><unexpected-request xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></unexpected-request></error>`,
			err: false,
		},
		3: {
			se:  stanza.Error{Type: stanza.Wait, Condition: stanza.UndefinedCondition},
			xml: `<error type="wait"><undefined-condition xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></undefined-condition></error>`,
			err: false,
		},
		4: {
			se:  stanza.Error{Type: stanza.Modify, By: jid.MustParse("test@example.net"), Condition: stanza.SubscriptionRequired},
			xml: `<error type="modify" by="test@example.net"><subscription-required xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></subscription-required></error>`,
			err: false,
		},
		5: {
			se:  stanza.Error{Type: stanza.Continue, Condition: stanza.ServiceUnavailable, Text: simpleText},
			xml: `<error type="continue"><service-unavailable xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></service-unavailable><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas">test</text></error>`,
			err: false,
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b, err := xml.Marshal(data.se)
			switch {
			case data.err && err == nil:
				t.Errorf("Expected an error when marshaling stanza error %v", data.se)
			case !data.err && err != nil:
				t.Error(err)
			case err != nil:
				return
			case string(b) != data.xml:
				t.Errorf("Expected marshaling stanza error '%v' to be:\n`%s`\nbut got:\n`%s`.", data.se, data.xml, string(b))
			}
		})
	}
}

func TestUnmarshalStanzaError(t *testing.T) {
	for i, data := range [...]struct {
		xml string
		se  stanza.Error
		err bool
	}{
		0: {"", stanza.Error{}, true},
		1: {`<error><unexpected-request xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></unexpected-request></error>`,
			stanza.Error{Condition: stanza.UnexpectedRequest}, false},
		2: {`<error type="cancel"><registration-required xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></registration-required></error>`,
			stanza.Error{Type: stanza.Cancel, Condition: stanza.RegistrationRequired}, false},
		3: {`<error type="cancel"><redirect xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></redirect></error>`,
			stanza.Error{Type: stanza.Cancel, Condition: stanza.Redirect}, false},
		4: {`<error type="wait"><undefined-condition xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></undefined-condition></error>`,
			stanza.Error{Type: stanza.Wait, Condition: stanza.UndefinedCondition}, false},
		5: {`<error type="modify" by="test@example.net"><subscription-required xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></subscription-required></error>`,
			stanza.Error{Type: stanza.Modify, By: jid.MustParse("test@example.net"), Condition: stanza.SubscriptionRequired}, false},
		6: {`<error type="continue"><service-unavailable xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></service-unavailable><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas">test</text></error>`,
			stanza.Error{Type: stanza.Continue, Condition: stanza.ServiceUnavailable, Text: simpleText}, false},
		7: {`<error type="auth"><resource-constraint xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></resource-constraint><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="en">test</text></error>`,
			stanza.Error{Type: stanza.Auth, Condition: stanza.ResourceConstraint, Text: map[string]string{
				"en": "test",
			}}, false},
		8: {`<error type="auth"><resource-constraint xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></resource-constraint><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="en">test</text><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="de">German</text></error>`,
			stanza.Error{Type: stanza.Auth, Condition: stanza.ResourceConstraint, Text: map[string]string{
				"en": "test",
				"de": "German",
			}}, false},
		9: {`<error type="auth"><remote-server-timeout xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></remote-server-timeout><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="en">test</text><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="es">Spanish</text></error>`,
			stanza.Error{Type: stanza.Auth, Condition: stanza.RemoteServerTimeout, Text: map[string]string{
				"en": "test",
				"es": "Spanish",
			}}, false},
		10: {`<error by=""><remote-server-not-found xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></remote-server-not-found></error>`,
			stanza.Error{By: jid.JID{}, Condition: stanza.RemoteServerNotFound}, false},
		11: {`<error><other xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></other></error>`,
			stanza.Error{Condition: stanza.Condition("other")}, false},
		12: {`<error><recipient-unavailable xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></recipient-unavailable><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="ac-u">test</text></error>`,
			stanza.Error{Condition: stanza.RecipientUnavailable, Text: map[string]string{
				"ac-u": "test",
			}}, false},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			se2 := stanza.Error{}
			err := xml.Unmarshal([]byte(data.xml), &se2)
			j1, j2 := data.se.By, se2.By
			data.se.By = jid.JID{}
			se2.By = jid.JID{}
			switch {
			case data.err && err == nil:
				t.Errorf("Expected an error when unmarshaling stanza error `%s`", data.xml)
			case !data.err && err != nil:
				t.Error(err)
			case err != nil:
				return
			case !j1.Equal(j2):
				t.Errorf(`Expected by="%v" but got by="%v"`, j1, j2)
			case !reflect.DeepEqual(data.se, se2):
				t.Errorf("Expected unmarshaled stanza error:\n`%#v`\nbut got:\n`%#v`", data.se, se2)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	stanzaErr := stanza.Error{Condition: stanza.RecipientUnavailable, Text: map[string]string{
		"ac-u": "test",
	}}
	r := stanzaErr.Wrap(xmlstream.Wrap(nil, xml.StartElement{Name: xml.Name{Local: "foo"}}))
	var buf strings.Builder
	e := xml.NewEncoder(&buf)
	_, err := xmlstream.Copy(e, r)
	if err != nil {
		t.Fatalf("error copying tokens: %v", err)
	}
	if err = e.Flush(); err != nil {
		t.Fatalf("error flushing buffer: %v", err)
	}
	const expected = `<error><recipient-unavailable xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"></recipient-unavailable><text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="ac-u">test</text><foo></foo></error>`
	if out := buf.String(); out != expected {
		t.Errorf("wrong output: want=%v, got=%v", expected, out)
	}
}
