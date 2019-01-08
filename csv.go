//
// Copyright (c) 2019 Open Collector, Inc.
//                    Moriyoshi Koizumi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package export

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type CSVRenderer struct {
	Out                 io.Writer
	FieldSeparator      []byte
	RecordSeparator     []byte
	QuoteCharacter      []byte
	QuoteCharacterStr   string
	CharsNeedToBeQuoted string
}

func (r *CSVRenderer) RenderHeader() error {
	typ := reflect.TypeOf(ExporterResult{})
	for i, nFields := 0, typ.NumField(); i < nFields; i++ {
		if i > 0 {
			_, err := r.Out.Write(r.FieldSeparator)
			if err != nil {
				return err
			}
		}
		_, err := r.Out.Write([]byte(typ.Field(i).Name))
		if err != nil {
			return err
		}
	}
	_, err := r.Out.Write(r.RecordSeparator)
	if err != nil {
		return err
	}
	return nil
}

func (r *CSVRenderer) RenderValue(v reflect.Value) string {
	vv := v.Interface()
	switch v := vv.(type) {
	case string:
		return v
	case time.Time:
		if v.IsZero() {
			return ""
		} else {
			return v.Format(time.RFC3339)
		}
	case int:
		return strconv.Itoa(v)
	default:
		return "N/A"
	}
}

func (r *CSVRenderer) Quote(v string) []byte {
	if strings.ContainsAny(v, r.CharsNeedToBeQuoted) {
		b := make([]byte, 0, len(v)+len(r.QuoteCharacter)*2)
		b = append(b, r.QuoteCharacter...)
		j := 0
		for j < len(v) {
			i := strings.Index(v[j:], r.QuoteCharacterStr)
			if i < 0 {
				break
			}
			b = append(b, v[j:j+i]...)
			b = append(b, r.QuoteCharacter...)
			b = append(b, r.QuoteCharacter...)
			j += i + len(r.QuoteCharacterStr)
		}
		b = append(b, v[j:]...)
		b = append(b, r.QuoteCharacter...)
		return b
	} else {
		return []byte(v)
	}
}

func (r *CSVRenderer) RenderOne(result *ExporterResult) error {
	v := reflect.ValueOf(result).Elem()
	for i, nFields := 0, v.NumField(); i < nFields; i++ {
		if i > 0 {
			_, err := r.Out.Write(r.FieldSeparator)
			if err != nil {
				return err
			}
		}
		_, err := r.Out.Write(r.Quote(r.RenderValue(v.Field(i))))
		if err != nil {
			return err
		}
	}
	_, err := r.Out.Write(r.RecordSeparator)
	if err != nil {
		return err
	}
	return nil
}

func (r *CSVRenderer) RenderPartialResults(results []ExporterResult) error {
	for i, _ := range results {
		err := r.RenderOne(&results[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CSVRenderer) RenderFooter() error {
	return nil
}

func (r *CSVRenderer) Fini() {}

func NewCSVRenderer(out io.Writer, fieldSeparator, recordSeparator, quoteCharacter string) *CSVRenderer {
	return &CSVRenderer{
		Out:                 out,
		FieldSeparator:      []byte(fieldSeparator),
		RecordSeparator:     []byte(recordSeparator),
		QuoteCharacter:      []byte(quoteCharacter),
		QuoteCharacterStr:   quoteCharacter,
		CharsNeedToBeQuoted: fieldSeparator + recordSeparator + quoteCharacter,
	}
}

func NewExcelCSVRenderer(out io.Writer) *CSVRenderer {
	return NewCSVRenderer(out, ",", "\r\n", "\"")
}
