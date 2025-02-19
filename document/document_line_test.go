package document

import (
	"testing"

	"github.com/campfhir/wsv/record"
)

func TestDocumentLineCompareStrings(t *testing.T) {
	line1 := documentLine{
		line: 1,
		fields: []record.RecordField{
			{FieldName: "key", Value: "OTS-1234", IsNull: false, FieldIndex: 0, RowIndex: 0, IsHeader: false},
		},
	}
	line2 := documentLine{
		line: 1,
		fields: []record.RecordField{
			{FieldName: "key", Value: "OTS-1235", IsNull: false, FieldIndex: 0, RowIndex: 0, IsHeader: false},
		},
	}
	a := line1.Compare("key", &line2, false)
	if a != -1 {
		t.Error("did not sort A correctly")
	}
	b := line2.Compare("key", &line1, false)
	if b != 1 {
		t.Error("did not sort B correctly")
	}
}

func TestDocumentLineCompareNumerics(t *testing.T) {
	line1 := documentLine{
		line: 1,
		fields: []record.RecordField{
			{FieldName: "key", Value: "1232", IsNull: false, FieldIndex: 0, RowIndex: 0, IsHeader: false},
		},
	}
	line2 := documentLine{
		line: 1,
		fields: []record.RecordField{
			{FieldName: "key", Value: "1231", IsNull: false, FieldIndex: 0, RowIndex: 0, IsHeader: false},
		},
	}
	a := line1.Compare("key", &line2, false)
	if a != 1 {
		t.Error("did not sort A correctly")
	}
	b := line2.Compare("key", &line1, false)
	if b != -1 {
		t.Error("did not sort B correctly")
	}
}
