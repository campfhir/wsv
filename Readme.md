# WSV

An implementation of **WSV (Whitespace Separated Values)** in Go, based on [Stenway/WSV-TS](https://github.com/Stenway/WSV-TS).

---

## Getting Started

Install the package:

```bash
go get github.com/campfhir/wsv
```

---

## Package Usage

### Reading a File

You can read a WSV file line by line using `Read()`, or read all lines at once with `ReadAll()`.

```go
package main

import (
    "os"
    "testing"

    wsv "github.com/campfhir/wsv/reader"
)

func TestRead(t *testing.T) {
    file, err := os.Open("testdata/sample.wsv")
    if err != nil {
        t.Fatal(err)
    }
    defer file.Close()

    r := wsv.NewReader(file)
    lines, err := r.ReadAll()
    if err != nil {
        t.Fatal(err)
    }

lineLoop:
    for _, line := range lines {
        for {
            field, err := line.NextField()
            if err == wsv.ErrEndOfLine {
                continue lineLoop
            }
            // Access field.Value, field.FieldName, or field.SerializeText()
        }
    }
}
```

---

### Writing a File

Documents can be written easily with the provided API:

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

   _ = line.Append("name")
   _ = line.Append("age")
   _ = line.Append("favorite color")

   err = doc.AppendLine(wsv.Field("scott"), wsv.NullField(), wsv.Field("red"))
   if err != nil {
      fmt.Println(err)
   }
}
```

---

## CLI Usage

A command-line tool is included for formatting and verifying WSV documents.

| Command/Option                    | Description                                                                                                                                              |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-i`, `-input`, `-f`, `-file`     | Input file (use `-` for stdin). Default: `-`.                                                                                                            |
| `-o`, `-output`                   | Output file (use `-` for stdout). Default: `-`.                                                                                                          |
| `-s`, `-sort`                     | Sort by column(s), separated by `;`. Use `::asc` or `::desc` to specify order. Default: ascending.                                                        |
| `-tabular`                        | Whether the document is tabular (each line has the same number of fields). Default: `true`.                                                               |
| `-v`, `-verify`                   | Verify that the input is valid WSV.                                                                                                                       |

---

## Marshal

The **Marshal** function converts Go structs into WSV-encoded output.

- Iterates over each element of a slice.
- Processes struct fields according to `wsv` tags.
- Supports custom formatting and comments.

### Struct Tag Format

```go
wsv:"[field name][,format:<fmt string>][,comment]"
```

- `field name`: If empty, defaults to the exported field’s name.
- `format:` specifies the format (uses `fmt.Sprintf` or `time.Format`).
- `comment` appends a comment to the line.

```go
type User struct {
  LastLogin time.Time `wsv:",comment"`
  Points    float64   `wsv:",format:%.4f,comment"`
}
```

---

### Supported Types

- `string`
- `int`
- `bool`
- `float`
- `time.Time`
- Any type implementing `MarshalWSV`

---

### String Fields

- `format:` attribute is ignored for `string`.

---

### Integer Fields

- Customizable with `format:` (default: `%d`).

```go
type Person struct {
   Age int `wsv:"age,format:%d"`
}
```

---

### Boolean Fields

- Format as `<true>|<false>` (default: `True|False`).

```go
type User struct {
  IsAdmin bool `wsv:"Admin,format:yes|no"`
}
```

---

### Float Fields

- Customizable with `format:` (default: `%.2f`).

```go
type Employee struct {
  Salary float32 `wsv:"Weekly Salary,format:%.2f"`
}
```

---

### Time Fields

- Customizable with `format:` (default: `time.RFC3339`).
- Accepts Go’s `time.Format` layouts or shorthand values:

  ```
  layout, ansic, unixdate, rubydate, rfc822, rfc822z,
  rfc850, rfc1123, rfc1123z, rfc3339, rfc3339nano,
  kitchen, stamp, stampmilli, stampmicro, stampnano,
  datetime, dateonly, date, timeonly, time
  ```

- Use single quotes `'` to escape commas:

```go
type TimeOff struct {
  Date      *time.Time `wsv:"PTO,format:'January 02, 2006'"`
  Requested time.Time  `wsv:"Requested,format:2006-01-02"`
  Approved  *time.Time `wsv:",format:rfc3339"`
}
```
