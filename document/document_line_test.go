package document_test

import (
	"testing"

	"github.com/campfhir/wsv/document"
)

func TestDocumentLineValidateWithTabularData(t *testing.T) {
	doc := document.NewDocument()
	line, _ := doc.AddLine()
	line.Append("First Name")
	line.Append("Fast Name")

	line, _ = doc.AddLine()
	line.Append("Scott")
	line.Append("Eremia-Roden")
	valid, err := line.Validate()
	if !valid {
		t.Error(err)
	}
}
