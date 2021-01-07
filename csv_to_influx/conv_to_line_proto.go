package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	protocol "github.com/influxdata/line-protocol"
)

type MyTag map[string]string

type MyMetric struct {
	TS      time.Time
	NameStr string
	Tags    []*protocol.Tag
	Fields  []*protocol.Field
}

var _ = protocol.Metric(&MyMetric{})

func convert_time(t string) (time.Time, error) {
	const long_form = "2006-01-02 15:04:05.999999"
	tim, err := time.Parse(long_form, t)
	if err != nil {
		return time.Now(), err
	}
	loc, err1 := time.LoadLocation("UTC")
	if err1 != nil {
		return time.Now(), err1
	}
	return tim.In(loc), nil
}

func main() {
	file_path := flag.String("path", "/data/file.csv", "path to data file in csv format")
	dump_path := flag.String("dump", "/data/dump", "path where to store line protocol files")
	bucket := flag.String("bucket", "gpu_monitoring", "bucket name")
	custom_tag := flag.String("tag", "job_name", "some custom tag to attach to the records")
	custom_tag_name := flag.String("tagname", "job", "some custom name for the tag")
	measurement := flag.String("measurement", "some_measurement", "some measurement name to give to the data")

	flag.Parse()

	fmt.Println("Importing file: " + *file_path)

	csv_file, err := os.Open(*file_path)
	if err != nil {
		fmt.Println("File not found!")
		return
	}
	reader := csv.NewReader(bufio.NewReader(csv_file))
	if reader == nil {
		panic(err)
	}
	defer csv_file.Close()

	// declare tags
	var tags []MyTag

	var metrics []string

	done_header := false

	for {
		// read line and check it
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			panic(error)
		}

		// check if we've done the header
		if !done_header {
			if line[0] != "__name__" &&
				line[0] != "controller_revision_hash" &&
				line[0] != "kubernetes_namespace" &&
				line[0] != "job" &&
				line[0] != "kubernetes_pod_name" &&
				line[0] != "name" &&
				line[0] != "pod_template_generation" &&
				line[0] != "" {

				// finished parsing multi-line header
				if line[0] == "time" {
					done_header = true
					if len(tags) < 1 {
						tags = append(tags, MyTag{*custom_tag_name: *custom_tag})
					} else {
						for k := 0; k < len(tags); k++ {
							tags[k][*custom_tag_name] = *custom_tag
						}
					}
				} else {
					for i := 0; i < len(line)/2; i++ {
						if len(tags) <= i {
							tags = append(tags, MyTag{line[i*2]: line[(i*2)+1]})
						} else {
							tags[i][line[i*2]] = line[(i*2)+1]
						}
					}
				}
			}
		} else {
			for j := 0; j < len(line)/2; j++ {
				if line[(j*2)+1] != "" && line[j*2] != "" {
					val, _ := strconv.ParseFloat(line[(j*2)+1], 64)
					t, _ := convert_time(line[j*2])
					var m *MyMetric = NewMetric(
						map[string]interface{}{"val": val},
						*measurement,
						tags[j],
						t)

					buf := &bytes.Buffer{}
					serializer := protocol.NewEncoder(buf)
					serializer.SetMaxLineBytes(4096)
					serializer.Encode(m)
					metrics = append(metrics, buf.String())
				}
			}
		}
	}

	write_line_proto(metrics, *bucket, *dump_path)

}

func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case bool, int64, string, float64:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return uint64(v)
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	default:
		panic("unsupported type")
	}
}

func (m *MyMetric) SortTags() {
	sort.Slice(m.Tags, func(i, j int) bool { return m.Tags[i].Key < m.Tags[j].Key })
}

func (m *MyMetric) SortFields() {
	sort.Slice(m.Fields, func(i, j int) bool { return m.Fields[i].Key < m.Fields[j].Key })
}
func NewMetric(fields map[string]interface{}, name string, tags map[string]string, ts time.Time) *MyMetric {
	m := &MyMetric{
		NameStr: name,
		Tags:    nil,
		Fields:  nil,
		TS:      ts,
	}
	if len(tags) > 0 {
		m.Tags = make([]*protocol.Tag, 0, len(tags))
		for k, v := range tags {
			m.Tags = append(m.Tags, &protocol.Tag{Key: k, Value: v})
		}
	}
	m.Fields = make([]*protocol.Field, 0, len(fields))
	for k, v := range fields {
		v := convertField(v)
		if v == nil {
			continue
		}
		m.Fields = append(m.Fields, &protocol.Field{Key: k, Value: v})
	}
	m.SortFields()
	m.SortTags()
	return m
}

// Name returns the name of the metric.
func (m *MyMetric) Name() string {
	return m.NameStr
}

// TagList returns a slice containing Tags of a Metric.
func (m *MyMetric) TagList() []*protocol.Tag {
	return m.Tags
}

// FieldList returns a slice containing the Fields of a Metric.
func (m *MyMetric) FieldList() []*protocol.Field {
	return m.Fields
}

// Time is the timestamp of a metric.
func (m *MyMetric) Time() time.Time {
	return m.TS
}

func write_line_proto(metrics []string, db string, base_path string) {
	u, err := uuid.NewRandom()
	f, err := os.Create(base_path + "/" + u.String() + ".txt")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(metrics); i++ {
		s := strings.TrimSpace(metrics[i])
		s = strings.Trim(s, "\n\r")
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		s = s + "\n"
		_, err := f.WriteString(s)
		if err != nil {
			f.Close()
			panic(err)
		}
	}
}
