package internal

import (
	"reflect"
	"strings"
	"time"
)

func ParseWSVTag(f reflect.StructField) (key string, isComment bool, format string, literalEmptyField bool) {
	key = f.Name
	if tag := f.Tag.Get("wsv"); tag != "" {
		parts := SplitQuoted(tag)
		if len(parts) > 0 {
			key = parts[0]
		}
		if key == "-" && len(parts) >= 2 {
			literalEmptyField = true
		}
		for _, p := range parts[1:] {
			switch {
			case strings.HasPrefix(p, "comment"):
				isComment = true
			case strings.HasPrefix(p, "format:"):
				format = strings.TrimPrefix(p, "format:")
			}
		}
	}
	return
}

// central lookup
var dateLayouts = map[string]string{
	"layout":      time.Layout,
	"ansic":       time.ANSIC,
	"unixdate":    time.UnixDate,
	"rubydate":    time.RubyDate,
	"rfc822":      time.RFC822,
	"rfc822z":     time.RFC822Z,
	"rfc850":      time.RFC850,
	"rfc1123":     time.RFC1123,
	"rfc1123z":    time.RFC1123Z,
	"rfc3339":     time.RFC3339,
	"rfc3339nano": time.RFC3339Nano,
	"kitchen":     time.Kitchen,
	"stamp":       time.Stamp,
	"stampmilli":  time.StampMilli,
	"stampmicro":  time.StampMicro,
	"stampnano":   time.StampNano,
	"datetime":    time.DateTime,
	"dateonly":    time.DateOnly,
	"date":        time.DateOnly,
	"timeonly":    time.TimeOnly,
	"time":        time.TimeOnly,
}

func ParseStructTagDateFormat(format string) string {
	if format == "" {
		return time.RFC3339
	}

	// normalize to lowercase
	key := strings.ToLower(format)

	if layout, ok := dateLayouts[key]; ok {
		return layout
	}

	// fallback: custom layout string
	return format
}
