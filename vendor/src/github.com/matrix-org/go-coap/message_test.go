package coap

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

// assertEqualMessages compares the e(xptected) message to the a(ctual) message
// and reports any diffs with t.Errorf.
func assertEqualMessages(t *testing.T, e, a Message) {
	if e.Type() != a.Type() {
		t.Errorf("Expected type %v, got %v", e.Type(), a.Type())
	}
	if e.Code() != a.Code() {
		t.Errorf("Expected code %v, got %v", e.Code(), a.Code())
	}
	if e.MessageID() != a.MessageID() {
		t.Errorf("Expected MessageID %v, got %v", e.MessageID(), a.MessageID())
	}
	if !bytes.Equal(e.Token(), a.Token()) {
		t.Errorf("Expected token %#v, got %#v", e.Token(), a.Token())
	}
	if !bytes.Equal(e.Payload(), a.Payload()) {
		t.Errorf("Expected payload %#v, got %#v", e.Payload(), a.Payload())
	}

	if len(e.AllOptions()) != len(a.AllOptions()) {
		t.Errorf("Expected %v options, got %v", e, a)
	} else {
		for i, _ := range e.AllOptions() {
			if e.AllOptions()[i].ID != a.AllOptions()[i].ID {
				t.Errorf("\nExpected option %v\n got %v", e.AllOptions()[i].ID, a.AllOptions()[i].ID)
				continue
			}
			switch e.AllOptions()[i].Value.(type) {
			case []byte:
				expected := e.AllOptions()[i].Value.([]byte)
				actual := a.AllOptions()[i].Value.([]byte)
				if !bytes.Equal(expected, actual) {
					t.Errorf("Expected Option ID %v value %v, got %v", e.AllOptions()[i].ID, expected, actual)
				}
			default:
				if e.AllOptions()[i].Value != a.AllOptions()[i].Value {
					t.Errorf("Expected Option ID %v value %v, got %v", e.AllOptions()[i].ID, e.AllOptions()[i].Value, a.AllOptions()[i].Value)
				}
			}
		}
	}
}

func TestMediaTypes(t *testing.T) {
	types := []interface{}{TextPlain, AppLinkFormat, AppXML, AppOctets, AppExi, AppJSON}
	exp := "coap.MediaType"
	for _, typ := range types {
		if got := fmt.Sprintf("%T", typ); got != exp {
			t.Errorf("Error on %#v, expected %q, was %q", typ, exp, got)
		}
	}
}

func TestOptionToBytes(t *testing.T) {
	tests := []struct {
		in  interface{}
		exp []byte
	}{
		{"", []byte{}},
		{[]byte{}, []byte{}},
		{"x", []byte{'x'}},
		{[]byte{'x'}, []byte{'x'}},
		{MediaType(3), []byte{0x3}},
		{3, []byte{0x3}},
		{838, []byte{0x3, 0x46}},
		{int32(838), []byte{0x3, 0x46}},
		{uint(838), []byte{0x3, 0x46}},
		{uint32(838), []byte{0x3, 0x46}},
	}

	for _, test := range tests {
		op := option{Value: test.in}
		buf := &bytes.Buffer{}
		err := op.writeData(buf)
		got := buf.Bytes()
		if err != nil {
			t.Error(err)
		} else if !bytes.Equal(test.exp, got) {
			t.Errorf("Error on %T(%v), got %#v, wanted %#v",
				test.in, test.in, got, test.exp)
		}
	}
}

func TestMessageConfirmable(t *testing.T) {
	tests := []struct {
		m   Message
		exp bool
	}{
		{&DgramMessage{MessageBase{typ: Confirmable}}, true},
		{&DgramMessage{MessageBase{typ: NonConfirmable}}, false},
	}

	for _, test := range tests {
		got := test.m.IsConfirmable()
		if got != test.exp {
			t.Errorf("Expected %v for %v", test.exp, test.m)
		}
	}
}

func TestMissingOption(t *testing.T) {
	got := (&DgramMessage{}).Option(MaxAge)
	if got != nil {
		t.Errorf("Expected nil, got %v", got)
	}
}

func TestOptionToBytesError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := option{Value: 3.1415926535897}.writeData(buf)
	if err == nil {
		t.Error("Expected panic. Didn't")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

func TestTypeString(t *testing.T) {
	tests := map[COAPType]string{
		Confirmable:    "Confirmable",
		NonConfirmable: "NonConfirmable",
		255:            "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v",
				code, code, exp)
		}
	}
}

func TestCodeString(t *testing.T) {
	tests := map[COAPCode]string{
		0:             "Unknown (0x0)",
		GET:           "GET",
		POST:          "POST",
		NotAcceptable: "NotAcceptable",
		255:           "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v",
				code, code, exp)
		}
	}
}

func TestEncodeMessageWithoutOptionsAndPayload(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{0x40, 0x1, 0x30, 0x39}
	if !bytes.Equal(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestEncodeMessageSmall(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}

	req.AddOption(ETag, []byte("weetag"))
	req.AddOption(MaxAge, 3)

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestEncodeMessageSmallWithPayload(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
			payload:   []byte("hi"),
		},
	}

	req.AddOption(ETag, []byte("weetag"))
	req.AddOption(MaxAge, 3)

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
		0xff, 'h', 'i',
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestInvalidMessageParsing(t *testing.T) {
	var invalidPackets = [][]byte{
		nil,
		{0x40},
		{0x40, 0},
		{0x40, 0, 0},
		{0xff, 0, 0, 0, 0, 0},
		{0x4f, 0, 0, 0, 0, 0},
		{0x45, 0, 0, 0, 0, 0},                // TKL=5 but packet is truncated
		{0x40, 0x01, 0x30, 0x39, 0x4d},       // Extended word length but no extra length byte
		{0x40, 0x01, 0x30, 0x39, 0x4e, 0x01}, // Extended word length but no full extra length word
	}

	for _, data := range invalidPackets {
		msg, err := ParseDgramMessage(data)
		if err == nil {
			t.Errorf("Unexpected success parsing short message (%#v): %v", data, msg)
		}
	}
}

func TestOptionsWithIllegalLengthAreIgnoredDuringParsing(t *testing.T) {
	exp := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 0xabcd,
			payload:   []byte{},
		},
	}
	msg, err := ParseDgramMessage([]byte{0x40, 0x01, 0xab, 0xcd,
		0x73, // URI-Port option (uint) with length 3 (valid lengths are 0-2)
		0x11, 0x22, 0x33, 0xff})
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}
	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", msg) {
		t.Errorf("Expected\n%#v\ngot\n%#v", exp, msg)
	}

	msg, err = ParseDgramMessage([]byte{0x40, 0x01, 0xab, 0xcd,
		0xd5, 0x01, // Max-Age option (uint) with length 5 (valid lengths are 0-4)
		0x11, 0x22, 0x33, 0x44, 0x55, 0xff})
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}
	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", msg) {
		t.Errorf("Expected\n%#v\ngot\n%#v", exp, msg)
	}
}

func TestDecodeMessageWithoutOptionsAndPayload(t *testing.T) {
	input := []byte{0x40, 0x1, 0x30, 0x39}
	msg, err := ParseDgramMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type() != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type())
	}
	if msg.Code() != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code())
	}
	if msg.MessageID() != 12345 {
		t.Errorf("Expected message ID 12345, got %v", msg.MessageID())
	}
	if len(msg.Token()) != 0 {
		t.Errorf("Incorrect token: %q", msg.Token())
	}
	if len(msg.Payload()) != 0 {
		t.Errorf("Incorrect payload: %q", msg.Payload())
	}
}

func TestDecodeMessageSmallWithPayload(t *testing.T) {
	input := []byte{
		0x40, 0x1, 0x30, 0x39, 0x21, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
		0xff, 'h', 'i',
	}

	msg, err := ParseDgramMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type() != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type())
	}
	if msg.Code() != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code())
	}
	if msg.MessageID() != 12345 {
		t.Errorf("Expected message ID 12345, got %v", msg.MessageID())
	}

	if !bytes.Equal(msg.Payload(), []byte("hi")) {
		t.Errorf("Incorrect payload: %q", msg.Payload())
	}
}

func TestEncodeMessageVerySmall(t *testing.T) {
	req := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	req.SetPathString("x")

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xb1, 0x78,
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

// Same as above, but with a leading slash
func TestEncodeMessageVerySmall2(t *testing.T) {
	req := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	req.SetPathString("/x")

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xb1, 0x78,
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestEncodeSeveral(t *testing.T) {
	tests := map[string][]string{
		"a":   []string{"a"},
		"axe": []string{"axe"},
		"a/b/c/d/e/f/h/g/i/j": []string{"a", "b", "c", "d", "e",
			"f", "h", "g", "i", "j"},
	}
	for p, a := range tests {
		m := &DgramMessage{
			MessageBase{
				typ:       Confirmable,
				code:      GET,
				messageID: 12345,
			},
		}
		m.SetPathString(p)
		buf := &bytes.Buffer{}
		err := m.MarshalBinary(buf)
		if err != nil {
			t.Errorf("Error encoding %#v", p)
			t.Fail()
			continue
		}
		m2, err := ParseDgramMessage(buf.Bytes())
		if err != nil {
			t.Fatalf("Can't parse my own message at %#v: %v", p, err)
		}

		if !reflect.DeepEqual(m2.Path(), a) {
			t.Errorf("Expected %#v, got %#v", a, m2.Path())
			t.Fail()
		}
	}
}

func TestPathAsOption(t *testing.T) {
	m := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	m.SetOption(LocationPath, []string{"a", "b"})
	buf := &bytes.Buffer{}
	err := m.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error marshaling: %v", err)
	}
	exp := []byte{0x40, 0x1, 0x30, 0x39, 0x81, 0x61, 0x1, 0x62}
	if !bytes.Equal(buf.Bytes(), exp) {
		t.Errorf("Got %#v, wanted %#v", buf.Bytes(), exp)
	}
}

func TestEncodePath14(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	req.SetPathString("123456789ABCDE")

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x01, // extended option length
		'1', '2', '3', '4', '5', '6', '7', '8',
		'9', 'A', 'B', 'C', 'D', 'E',
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestEncodePath15(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	req.SetPathString("123456789ABCDEF")

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x02, // extended option length
		'1', '2', '3', '4', '5', '6', '7', '8',
		'9', 'A', 'B', 'C', 'D', 'E', 'F',
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestEncodeLargePath(t *testing.T) {
	req := DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
		},
	}
	req.SetPathString("this_path_is_longer_than_fifteen_bytes")

	if req.PathString() != "this_path_is_longer_than_fifteen_bytes" {
		t.Fatalf("Didn't get back the same path I posted: %v",
			req.PathString())
	}

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	// Inspected by hand.
	exp := []byte{
		// extended length           0x19 + 13 = 38
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x19, 0x74, 0x68, 0x69,
		0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74,
		0x68, 0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65,
		0x65, 0x6e, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}
	if !reflect.DeepEqual(exp, buf.Bytes()) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, buf.Bytes())
	}
}

func TestDecodeLargePath(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x19, 0x74, 0x68,
		0x69, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74, 0x68,
		0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65, 0x65, 0x6e,
		0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}

	req, err := ParseDgramMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	path := "this_path_is_longer_than_fifteen_bytes"

	exp := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
			payload:   []byte{},
		},
	}

	exp.SetOption(URIPath, path)

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		buf := &bytes.Buffer{}
		exp.MarshalBinary(buf)
		t.Fatalf("Expected\n%#v\ngot\n%#v\nfor %#v", exp, req, buf.Bytes())
	}
}

func TestDecodeMessageSmaller(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
	}

	req, err := ParseDgramMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	exp := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
			payload:   []byte{},
		},
	}

	exp.SetOption(ETag, []byte("weetag"))
	exp.SetOption(MaxAge, uint32(3))

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, req)
	}
}

func TestByteEncoding(t *testing.T) {
	tests := []struct {
		Value    uint32
		Expected []byte
	}{
		{0, nil},
		{13, []byte{13}},
		{1024, []byte{4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		buf := &bytes.Buffer{}
		err := encodeInt(buf, v.Value)
		if err != nil {
			t.Error(err)
		}
		got := buf.Bytes()
		if !reflect.DeepEqual(got, v.Expected) {
			t.Fatalf("Expected %#v, got %#v for %v",
				v.Expected, got, v.Value)
		}
	}
}

func TestByteDecoding(t *testing.T) {
	tests := []struct {
		Value uint32
		Bytes []byte
	}{
		{0, nil},
		{0, []byte{0}},
		{0, []byte{0, 0}},
		{0, []byte{0, 0, 0}},
		{0, []byte{0, 0, 0, 0}},
		{13, []byte{13}},
		{13, []byte{0, 13}},
		{13, []byte{0, 0, 13}},
		{13, []byte{0, 0, 0, 13}},
		{1024, []byte{4, 0}},
		{1024, []byte{4, 0}},
		{1024, []byte{0, 4, 0}},
		{1024, []byte{0, 0, 4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{984284, []byte{0, 0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		got := decodeInt(v.Bytes)
		if v.Value != got {
			t.Fatalf("Expected %v, got %v for %#v",
				v.Value, got, v.Bytes)
		}
	}
}

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 0 |   0   |     GET=1     |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  11   |  11   |      "temperature" (11 B) ...                 |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1(t *testing.T) {
	input := append([]byte{0x40, 1, 0x7d, 0x34,
		(11 << 4) | 11}, []byte("temperature")...)

	msg, err := ParseDgramMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type() != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type())
	}
	if msg.Code() != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code())
	}
	if msg.MessageID() != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID())
	}

	if msg.Option(URIPath).(string) != "temperature" {
		t.Errorf("Incorrect uri path: %q", msg.Option(URIPath))
	}

	if len(msg.Token()) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token())
	}
	if len(msg.Payload()) > 0 {
		t.Errorf("Incorrect payload: %q", msg.Payload())
	}
}

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 2 |   0   |    2.05=69    |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |1 1 1 1 1 1 1 1|      "22.3 C" (6 B) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1Res(t *testing.T) {
	input := append([]byte{0x60, 69, 0x7d, 0x34, 0xff},
		[]byte("22.3 C")...)

	msg, err := ParseDgramMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type() != Acknowledgement {
		t.Errorf("Expected message type confirmable, got %v", msg.Type())
	}
	if msg.Code() != Content {
		t.Errorf("Expected message code Content, got %v", msg.Code())
	}
	if msg.MessageID() != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID())
	}

	if len(msg.Token()) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token())
	}
	if !bytes.Equal(msg.Payload(), []byte("22.3 C")) {
		t.Errorf("Incorrect payload: %q", msg.Payload())
	}
}

func TestIssue15(t *testing.T) {

	input := []byte{0x53, 0x2, 0x7a,
		0x23, 0x1, 0x2, 0x3, 0xb1, 0x45, 0xd, 0xd, 0x73, 0x70, 0x61,
		0x72, 0x6b, 0x2f, 0x63, 0x63, 0x33, 0x30, 0x30, 0x30, 0x2d,
		0x70, 0x61, 0x74, 0x63, 0x68, 0x2d, 0x76, 0x65, 0x72, 0x73,
		0x69, 0x6f, 0x6e, 0xff, 0x31, 0x2e, 0x32, 0x38}
	msg, err := ParseDgramMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if !bytes.Equal(msg.Token(), []byte{1, 2, 3}) {
		t.Errorf("Expected token = [1, 2, 3], got %v", msg.Token())
	}

	if !bytes.Equal(msg.Payload(), []byte{0x31, 0x2e, 0x32, 0x38}) {
		t.Errorf("Expected payload = {0x31, 0x2e, 0x32, 0x38}, got %v", msg.Payload())
	}

	pathExp := "E/spark/cc3000-patch-version"
	if got := msg.PathString(); got != pathExp {
		t.Errorf("Expected path %q, got %q", pathExp, got)
	}
}

func TestErrorOptionMarker(t *testing.T) {
	input := []byte{0x53, 0x2, 0x7a, 0x23,
		0x1, 0x2, 0x3, 0xbf, 0x01, 0x02, 0x03, 0x04, 0x05, 0x6, 0x7, 0x8, 0x9,
		0xa, 0xb, 0xc, 0xe, 0xf, 0x10}
	msg, err := ParseDgramMessage(input)
	if err == nil {
		t.Errorf("Unexpected success parsing malformed option: %v", msg)
	}
}

func TestDecodeContentFormatOptionToMediaType(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0xc1, 0x32, 0x51, 0x29,
	}

	parsedMsg, err := ParseDgramMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	expected := "coap.MediaType"
	actualContentFormatType := fmt.Sprintf("%T", parsedMsg.Option(ContentFormat))
	if expected != actualContentFormatType {
		t.Fatalf("Expected %#v got %#v", expected, actualContentFormatType)
	}
	actualAcceptType := fmt.Sprintf("%T", parsedMsg.Option(Accept))
	if expected != actualAcceptType {
		t.Fatalf("Expected %#v got %#v", expected, actualAcceptType)
	}
}

func TestEncodeMessageWithAllOptions(t *testing.T) {
	req := &DgramMessage{
		MessageBase{
			typ:       Confirmable,
			code:      GET,
			messageID: 12345,
			token:     []byte("TOKEN"),
			payload:   []byte("PAYLOAD"),
		},
	}

	req.AddOption(IfMatch, []byte("IFMATCH"))
	req.AddOption(URIHost, "URIHOST")
	req.AddOption(ETag, []byte("ETAG"))
	req.AddOption(IfNoneMatch, []byte{})
	req.AddOption(Observe, uint32(9999))
	req.AddOption(URIPort, uint32(5683))
	req.AddOption(LocationPath, "LOCATIONPATH")
	req.AddOption(URIPath, "URIPATH")
	req.AddOption(ContentFormat, TextPlain)
	req.AddOption(MaxAge, uint32(9999))
	req.AddOption(URIQuery, "URIQUERY")
	req.AddOption(Accept, TextPlain)
	req.AddOption(LocationQuery, "LOCATIONQUERY")
	req.AddOption(ProxyURI, "PROXYURI")
	req.AddOption(ProxyScheme, "PROXYSCHEME")
	req.AddOption(Size1, uint32(9999))
	req.AddOption(Block1, uint32(66560))
	req.AddOption(Size2, uint32(9999))
	req.AddOption(Block2, uint32(66560))

	buf := &bytes.Buffer{}
	err := req.MarshalBinary(buf)
	if err != nil {
		t.Fatalf("Error encoding request: %v", err)
	}

	parsedMsg, err := ParseDgramMessage(buf.Bytes())
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}
	assertEqualMessages(t, req, parsedMsg)
}

func TestBlockWiseTransfer(t *testing.T) {
	peer0_0 := []byte{
		0x58, 0x01, 0x80, 0x5e, 0x41, 0x37, 0x2f, 0x10,
		0x4e, 0xab, 0xef, 0xfc, 0xbd, 0x04, 0x61, 0x6c,
		0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x4d,
		0x05, 0x69, 0x66, 0x3d, 0x6f, 0x69, 0x63, 0x2e,
		0x69, 0x66, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x6c,
		0x69, 0x6e, 0x65, 0x0d, 0x1f, 0x72, 0x74, 0x3d,
		0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x6b, 0x69,
		0x73, 0x74, 0x6c, 0x65, 0x72, 0x2e, 0x6b, 0x69,
		0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e,
		0x61, 0x6c, 0x6c, 0x61, 0x74, 0x74, 0x72, 0x69,
		0x62, 0x75, 0x74, 0x65, 0x74, 0x79, 0x70, 0x65,
		0x73, 0x22, 0x27, 0x10, 0xe2, 0x06, 0xe3, 0x08,
		0x00, 0xe1, 0xf6, 0xe6, 0xc0}

	peer0_1 := []byte{
		0x58, 0x01, 0x6e, 0x03, 0x41, 0x37, 0x2f, 0x10,
		0x4e, 0xab, 0xef, 0xfc, 0xbd, 0x04, 0x61, 0x6c,
		0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x4d,
		0x05, 0x69, 0x66, 0x3d, 0x6f, 0x69, 0x63, 0x2e,
		0x69, 0x66, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x6c,
		0x69, 0x6e, 0x65, 0x0d, 0x1f, 0x72, 0x74, 0x3d,
		0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x6b, 0x69,
		0x73, 0x74, 0x6c, 0x65, 0x72, 0x2e, 0x6b, 0x69,
		0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e,
		0x61, 0x6c, 0x6c, 0x61, 0x74, 0x74, 0x72, 0x69,
		0x62, 0x75, 0x74, 0x65, 0x74, 0x79, 0x70, 0x65,
		0x73, 0x22, 0x27, 0x10, 0x61, 0x16, 0xe2, 0x06,
		0xdd, 0x08, 0x00, 0xe1, 0xf6, 0xe6, 0xc0}

	peer1_0 := []byte{
		0x58, 0x45, 0x80, 0x5e, 0x41, 0x37, 0x2f, 0x10,
		0x4e, 0xab, 0xef, 0xfc, 0xbd, 0x04, 0x61, 0x6c,
		0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x12,
		0x27, 0x10, 0xb1, 0x0e, 0x52, 0x05, 0x28, 0xe2,
		0x06, 0xdc, 0x08, 0x00, 0xe1, 0xf6, 0xe2, 0xc0,
		0xff, 0xbf, 0x62, 0x72, 0x74, 0x81, 0x78, 0x29,
		0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x6b, 0x69,
		0x73, 0x74, 0x6c, 0x65, 0x72, 0x2e, 0x6b, 0x69,
		0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e,
		0x61, 0x6c, 0x6c, 0x61, 0x74, 0x74, 0x72, 0x69,
		0x62, 0x75, 0x74, 0x65, 0x74, 0x79, 0x70, 0x65,
		0x73, 0x62, 0x69, 0x66, 0x84, 0x68, 0x6f, 0x69,
		0x63, 0x2e, 0x69, 0x66, 0x2e, 0x72, 0x69, 0x6f,
		0x69, 0x63, 0x2e, 0x69, 0x66, 0x2e, 0x72, 0x77,
		0x64, 0x74, 0x65, 0x73, 0x74, 0x6f, 0x6f, 0x69,
		0x63, 0x2e, 0x69, 0x66, 0x2e, 0x62, 0x61, 0x73,
		0x65, 0x6c, 0x69, 0x6e, 0x65, 0x6f, 0x62, 0x69,
		0x6e, 0x61, 0x72, 0x79, 0x41, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x42, 0xca, 0xfe,
		0x6d, 0x62, 0x6f, 0x6f, 0x6c, 0x41, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0xf5, 0x6f,
		0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x41, 0x74,
		0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0xfb,
		0x40, 0x45, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x6c, 0x69, 0x6e, 0x74, 0x41, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x18, 0x2a, 0x6d,
		0x6e, 0x75, 0x6c, 0x6c, 0x41, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0xf6, 0x71, 0x72,
		0x61, 0x6e, 0x67, 0x65, 0x49, 0x6e, 0x74, 0x41,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x18, 0x32, 0x77, 0x72, 0x65, 0x61, 0x64, 0x4f,
		0x6e, 0x6c, 0x79, 0x53, 0x74, 0x72, 0x69, 0x6e,
		0x67, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x62, 0x72, 0x6f, 0x77, 0x72, 0x65,
		0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x74, 0x61,
		0x74, 0x69, 0x6f, 0x6e, 0x41, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0xbf, 0x69, 0x61,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x65, 0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0x6f,
		0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x41, 0x74,
		0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x6b,
		0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x76, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x31, 0x62, 0x69, 0x6e, 0x61, 0x72,
		0x79, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x82, 0x42, 0xca, 0xfe, 0x42, 0xca,
		0xfe, 0x74, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
		0x31, 0x62, 0x6f, 0x6f, 0x6c, 0x41, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82, 0xf5,
		0xf4, 0x76, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
		0x31, 0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x41,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x82, 0xfb, 0x3f, 0xf1, 0x99, 0x99, 0x99, 0x99,
		0x99, 0x9a, 0xfb, 0x3f, 0xf3, 0x33, 0x33, 0x33,
		0x33, 0x33, 0x33, 0x73, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x31, 0x69, 0x6e, 0x74, 0x41, 0x74,
		0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82,
		0x01, 0x02, 0x78, 0x1e, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x31, 0x72, 0x65, 0x70, 0x72, 0x65,
		0x73, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f,
		0x6e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x82, 0xbf, 0x69, 0x61, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x65, 0x76,
		0x61, 0x6c, 0x75, 0x65, 0xff, 0xbf, 0x69, 0x61,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x65, 0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0x76,
		0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x31, 0x73,
		0x74, 0x72, 0x69, 0x6e, 0x67, 0x41, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82, 0x65,
		0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x65, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x76, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x32, 0x62, 0x69, 0x6e, 0x61, 0x72,
		0x79, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x82, 0x82, 0x42, 0xca, 0xfe, 0x42,
		0xca, 0xfe, 0x82, 0x42, 0xca, 0xfe, 0x42, 0xca,
		0xfe, 0x74, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
		0x32, 0x62, 0x6f, 0x6f, 0x6c, 0x41, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82, 0x82,
		0xf5, 0xf4, 0x82, 0xf4, 0xf5, 0x76, 0x76, 0x65,
		0x63, 0x74, 0x6f, 0x72, 0x32, 0x64, 0x6f, 0x75,
		0x62, 0x6c, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69,
		0x62, 0x75, 0x74, 0x65, 0x82, 0x82, 0xfb, 0x3f,
		0xf1, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a, 0xfb,
		0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
		0x82, 0xfb, 0x40, 0x00, 0xcc, 0xcc, 0xcc, 0xcc,
		0xcc, 0xcd, 0xfb, 0x40, 0x01, 0x99, 0x99, 0x99,
		0x99, 0x99, 0x9a, 0x73, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x32, 0x69, 0x6e, 0x74, 0x41, 0x74,
		0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82,
		0x82, 0x01, 0x02, 0x82, 0x03, 0x04, 0x78, 0x1e,
		0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x32, 0x72,
		0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x74,
		0x61, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x74, 0x74,
		0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x82, 0x82,
		0xbf, 0x69, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62,
		0x75, 0x74, 0x65, 0x65, 0x76, 0x61, 0x6c, 0x75,
		0x65, 0xff, 0xbf, 0x69, 0x61, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x65, 0x76, 0x61,
		0x6c, 0x75, 0x65, 0xff, 0x82, 0xbf, 0x69, 0x61,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x65, 0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0xbf,
		0x69, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x65,
		0xff, 0x76, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
		0x32, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x41,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x82, 0x82, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f,
		0x65, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x82, 0x65,
		0x48, 0x61, 0x6c, 0x6c, 0x6f, 0x64, 0x77, 0x65,
		0x6c, 0x74, 0x76, 0x76, 0x65, 0x63, 0x74, 0x6f,
		0x72, 0x33, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79,
		0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
		0x65, 0x82, 0x82, 0x82, 0x42, 0xca, 0xfe, 0x42,
		0xca, 0xfe, 0x82, 0x42, 0xca, 0xfe, 0x42, 0xca,
		0xfe, 0x82, 0x82, 0x42, 0xca, 0xfe, 0x42, 0xca,
		0xfe, 0x82, 0x42, 0xca, 0xfe, 0x42, 0xca, 0xfe,
		0x74, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x33,
		0x62, 0x6f, 0x6f, 0x6c, 0x41, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x82, 0x82, 0x82,
		0xf5, 0xf4, 0x82, 0xf4, 0xf5, 0x82, 0x82, 0xf5,
		0xf4, 0x82, 0xf4, 0xf5, 0x76, 0x76, 0x65, 0x63,
		0x74, 0x6f, 0x72, 0x33, 0x64, 0x6f, 0x75, 0x62,
		0x6c, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62,
		0x75, 0x74, 0x65, 0x82, 0x82, 0x82, 0xfb, 0x3f,
		0xf1, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a, 0xfb,
		0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33,
		0x82, 0xfb, 0x40, 0x00, 0xcc, 0xcc, 0xcc, 0xcc,
		0xcc, 0xcd, 0xfb, 0x40, 0x01, 0x99, 0x99, 0x99,
		0x99, 0x99, 0x9a, 0x82, 0x82, 0xfb, 0x40, 0x08,
		0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0xcd, 0xfb, 0x40,
		0x09, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a, 0x82,
		0xfb, 0x40, 0x10, 0x66, 0x66, 0x66, 0x66, 0x66,
		0x66, 0xfb, 0x40, 0x10, 0xcc, 0xcc, 0xcc, 0xcc,
		0xcc}

	peer1_1 := []byte{
		0x58, 0x45, 0x6e, 0x03, 0x41, 0x37, 0x2f, 0x10,
		0x4e, 0xab, 0xef, 0xfc, 0xbd, 0x04, 0x61, 0x6c,
		0x6c, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x12,
		0x27, 0x10, 0xb1, 0x16, 0xe2, 0x06, 0xe1, 0x08,
		0x00, 0xe1, 0xf6, 0xe2, 0xc0, 0xff, 0xcd, 0x73,
		0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x33, 0x69,
		0x6e, 0x74, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62,
		0x75, 0x74, 0x65, 0x82, 0x82, 0x82, 0x01, 0x02,
		0x82, 0x03, 0x04, 0x82, 0x82, 0x05, 0x06, 0x82,
		0x07, 0x08, 0x78, 0x1e, 0x76, 0x65, 0x63, 0x74,
		0x6f, 0x72, 0x33, 0x72, 0x65, 0x70, 0x72, 0x65,
		0x73, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f,
		0x6e, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x82, 0x82, 0x82, 0xbf, 0x69, 0x61,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x65, 0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0xbf,
		0x69, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x65,
		0xff, 0x82, 0xbf, 0x69, 0x61, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x65, 0x76, 0x61,
		0x6c, 0x75, 0x65, 0xff, 0xbf, 0x69, 0x61, 0x74,
		0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x65,
		0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0x82, 0x82,
		0xbf, 0x69, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62,
		0x75, 0x74, 0x65, 0x65, 0x76, 0x61, 0x6c, 0x75,
		0x65, 0xff, 0xbf, 0x69, 0x61, 0x74, 0x74, 0x72,
		0x69, 0x62, 0x75, 0x74, 0x65, 0x65, 0x76, 0x61,
		0x6c, 0x75, 0x65, 0xff, 0x82, 0xbf, 0x69, 0x61,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x65, 0x76, 0x61, 0x6c, 0x75, 0x65, 0xff, 0xbf,
		0x69, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75,
		0x74, 0x65, 0x65, 0x76, 0x61, 0x6c, 0x75, 0x65,
		0xff, 0x76, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72,
		0x33, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x41,
		0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
		0x82, 0x82, 0x82, 0x65, 0x48, 0x65, 0x6c, 0x6c,
		0x6f, 0x65, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x82,
		0x65, 0x48, 0x61, 0x6c, 0x6c, 0x6f, 0x64, 0x77,
		0x65, 0x6c, 0x74, 0x82, 0x82, 0x65, 0x48, 0x65,
		0x6c, 0x6c, 0x6f, 0x65, 0x77, 0x6f, 0x72, 0x6c,
		0x64, 0x82, 0x65, 0x48, 0x61, 0x6c, 0x6c, 0x6f,
		0x64, 0x77, 0x65, 0x6c, 0x74, 0xff}

	_, err := ParseDgramMessage(peer0_0)
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}

	peer0_1Msg, err := ParseDgramMessage(peer0_1)
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}

	peer1_0Msg, err := ParseDgramMessage(peer1_0)
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}

	peer1_1Msg, err := ParseDgramMessage(peer1_1)
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}

	if peer1_0Msg.Option(Block2).(uint32) != 14 {
		t.Fatalf("peer1_0Msg.Option(Block2): %v", peer1_0Msg.Option(Block2).(uint32))
	}

	if peer0_1Msg.Option(Block2).(uint32) != 22 {
		t.Fatalf("peer0_1Msg.Option(Block2): %v", peer0_1Msg.Option(Block2).(uint32))
	}

	if peer1_1Msg.Option(Block2).(uint32) != 22 {
		t.Fatalf("peer1_1Msg.Option(Block2): %v", peer1_1Msg.Option(Block2).(uint32))
	}
}