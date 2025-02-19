# WSV

This is an implementation of WSV in Go as described by [https://github.com/Stenway/WSV-TS](https://github.com/Stenway/WSV-TS).

## Getting Started

```bash
go get github.com/campfhir/wsv
```

## Reading Usage

When using the reader you can read line by line, `Read()` or use the convenient `ReadAll()` function which reads all lines into a slice of rows of records/fields.

```go
package main

import (
    "fmt"
    "os"
    "testing"

    wsv "github.com/campfhir/wsv/reader"
)

func TestRead(t *testing.T) {
    dir, ok := os.LookupEnv("PROJECT_DIR")
    if !ok {
        t.Error("PROJECT_DIR env not FOUND")
        t.FailNow()
        return
    }
    file, err := os.Open("testdata/sample.wsv")
    if err != nil {
        t.Error(err)
        t.FailNow()
        return
    }
    r := wsv.NewReader(file)
    lines, err := r.ReadAll()
    if err != nil {
        t.Error(err)
        return
    }
lineLoop:
    for _, line := range lines {
        for {
            // field
            field, err := line.NextField()
            if err == wsv.ErrEndOfLine {
                continue lineLoop
            }
            // field.SerializeText()
            // field.Value
            // field.FieldName
        }
    }
}

```

## Writing Usage

When writing a document can be done with a few APIs. Below is a sample application.

```go
package main

import (
   "fmt"

   wsv "github.com/campfhir/wsv/document"
)

func main() {
   doc := wsv.NewDocument()
   line, err := doc.AddLine()
   if err != nil {
      fmt.Println(err)
      return
   }
   err = line.Append("name")
   if err != nil {
      fmt.Println(err)
      return
   }
   err = line.Append("age")
   if err != nil {
      fmt.Println(err)
      return
   }
   err = line.Append("favorite color")
   if err != nil {
      fmt.Println(err)
      return
   }
   err = doc.AppendLine(wsv.Field("scott"), wsv.NullField(), wsv.Field("red"))
   if err != nil {
      fmt.Println(err)
      return
   }
}
```
