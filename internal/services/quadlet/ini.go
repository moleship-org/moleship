package quadlet

import (
	"fmt"
	"sort"
	"strings"
)

type iniBuilder struct {
	sb strings.Builder
}

func newIniBuilder() *iniBuilder {
	return &iniBuilder{}
}

func (b *iniBuilder) section(name string) *iniBuilder {
	if b.sb.Len() > 0 {
		b.sb.WriteString("\n")
	}
	b.sb.WriteString("[" + name + "]\n")
	return b
}

func (b *iniBuilder) kv(key, value string) *iniBuilder {
	if value == "" {
		return b
	}
	fmt.Fprintf(&b.sb, "%s=%s\n", key, value)
	return b
}

func (b *iniBuilder) kvList(key string, values []string) *iniBuilder {
	for _, v := range values {
		if v == "" {
			continue
		}
		fmt.Fprintf(&b.sb, "%s=%s\n", key, v)
	}
	return b
}

func (b *iniBuilder) kvSpaceJoined(key string, values []string) *iniBuilder {
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	if len(nonEmpty) == 0 {
		return b
	}
	fmt.Fprintf(&b.sb, "%s=%s\n", key, strings.Join(nonEmpty, " "))
	return b
}

func (b *iniBuilder) kvMap(key string, values map[string]string) *iniBuilder {
	if len(values) == 0 {
		return b
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b.sb, "%s=%s=%s\n", key, name, quoteIfNeeded(values[name]))
	}
	return b
}

func (b *iniBuilder) String() string {
	return b.sb.String()
}

func quoteIfNeeded(value string) string {
	if value == "" {
		return `""`
	}

	needsQuoting := strings.ContainsAny(value, " \t\"'\\$")
	if !needsQuoting {
		return value
	}

	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"', '\\':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

func quoteArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = quoteIfNeeded(a)
	}
	return strings.Join(quoted, " ")
}
