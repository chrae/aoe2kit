package xs

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestInspectDataBytesAutoV9Payload(t *testing.T) {
	data := v9Payload(t)
	report := InspectDataBytes(data, DataInspectOptions{})
	if !report.OK {
		t.Fatalf("report not OK: %#v", report.Errors)
	}
	if report.Mode != "auto_heuristic" {
		t.Fatalf("mode = %q", report.Mode)
	}
	if len(report.Values) != 10 {
		t.Fatalf("values = %d, want 10", len(report.Values))
	}
	if report.Values[0].String != "A2K_RTV9_PERSIST" {
		t.Fatalf("magic = %q", report.Values[0].String)
	}
	if report.Values[3].Int == nil || *report.Values[3].Int != 424242 {
		t.Fatalf("sentinel = %#v", report.Values[3].Int)
	}
	if report.Values[8].String != "end" {
		t.Fatalf("footer label = %q", report.Values[8].String)
	}
}

func TestInspectDataBytesTypedV9Payload(t *testing.T) {
	types, err := ParseDataTypes("string,int,string,int,string,int,int,int,string,int")
	if err != nil {
		t.Fatal(err)
	}
	report := InspectDataBytes(v9Payload(t), DataInspectOptions{Types: types})
	if !report.OK {
		t.Fatalf("report not OK: %#v", report.Errors)
	}
	if report.Mode != "typed" {
		t.Fatalf("mode = %q", report.Mode)
	}
	if report.Values[7].Int == nil || *report.Values[7].Int != 123456 {
		t.Fatalf("xp = %#v", report.Values[7].Int)
	}
	if report.RemainingBytes != 0 {
		t.Fatalf("remaining = %d", report.RemainingBytes)
	}
}

func v9Payload(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	writeString := func(s string) {
		if err := binary.Write(&b, binary.LittleEndian, uint32(len(s))); err != nil {
			t.Fatal(err)
		}
		b.WriteString(s)
	}
	writeInt := func(v int32) {
		if err := binary.Write(&b, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	writeString("A2K_RTV9_PERSIST")
	writeInt(9)
	writeString("writer_v9a")
	writeInt(424242)
	writeString("player")
	writeInt(3)
	writeInt(17)
	writeInt(123456)
	writeString("end")
	writeInt(9001)
	return b.Bytes()
}
