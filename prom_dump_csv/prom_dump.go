package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
    "log"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kr/pretty"
)

type InstanceDataResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

type InstanceDataResponse struct {
	ResultType string               `json:"resultType"`
	Result     []InstanceDataResult `json:"result"`
}

type InstanceResponse struct {
	Status string               `json:"status"`
	Data   InstanceDataResponse `json:"data"`
}

type MainDataResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

type MainDataResponse struct {
	ResultType string           `json:"resultType"`
	Result     []MainDataResult `json:"result"`
}

type MainResponse struct {
	Status string           `json:"status"`
	Data   MainDataResponse `json:"data"`
}

type ConfigJsonMeasurement struct {
	Name    string   `json:"name"`
    // Labels we want to keep in our csv output
    Labels  []string   `json:"labels"`
	Metrics []string `json:"metrics"`
}

type ConfigJson struct {
	Measurements []ConfigJsonMeasurement `json:"measurements"`
}

// Measurement is the name of the metric we want to query
type Measurement struct {
    Metrics []string
    Labels []string
}

// MetricInstanceCombination combines metrics string with a set of `instances` 
type MetricInstanceCombination struct {
	Metric    string
    // labels we care about can be empty
    Labels    []string
	Instances []string
}

// ParsedMetrics a collection of metrics parsed from prometheus
// used for writing data to csv
type ParsedMetrics struct {
    Metric string
    Instance string
    // same as metric?
    Job string
    Labels []string
    DataPoints []DataPoint
}

//GetHeader returns header array plust the idx of any labels
// if no extra labels than labelIdx will equal len(header)-1
// This includes all metrics for a single instance...
func (pm *ParsedMetrics) GetHeader() ([]string, int) {
    header := []string{"__name__", "instance", "time", "value"}
    labelIdx := len(header)
    keys := make([]string, 0, len(pm.Labels))
    for _, k := range pm.Labels {
        keys = append(keys, k)
    }
    // We sort here encase client hasnt done so. O(n) if sorted -- no exceptions if its not...
    sort.Strings(keys)
    header = append(header, keys...)
    return header, labelIdx
}

// GetRows returns array of parsed data as rows to be used in csv file.
// First arrays contains header data from `ParsedMetrics.GetHeader()`
func (pm *ParsedMetrics) GetRows() [][]string {
    header, labelIdx := pm.GetHeader()
    // pre-allocate array to length of rows + 1 for header
    rows := make([][]string, 0, len(pm.DataPoints) + 1)
    rows = append(rows, header)
    for _, dp := range pm.DataPoints {
        row := make([]string, 0, len(header))
        row = append(row, pm.Metric)
        row = append(row, pm.Instance)
        //row = append(row, pm.Job)
        row = append(row, conv_time_record(dp.Timestamp))
        row = append(row, dp.Value)
        // handle label values
        if labelIdx < len(header)-1 {
            for i:=labelIdx; i<len(header); i++ {
                // add labels to row in header order
                row = append(row, dp.Labels[header[i]])
            }
        }
        rows = append(rows, row)
    }
    return rows
}

// DataPoint encapsulates parsed metric data
type DataPoint struct {
    Timestamp string
    Value string
    Labels map[string]string
}

// get_config from path and a list of target groups
// this function itterates each measurement in the config
// and collect prometheus query strings
func get_config(group string, path string) []Measurement {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config ConfigJson
    err = json.Unmarshal([]byte(byteValue), &config)
    if err != nil {
        panic(fmt.Sprintf("Error parsing config: %v", err))
    }
	var measurements []Measurement
	var groups []string = strings.Split(group, ",")
	for i := 0; i < len(config.Measurements); i++ {
		for j := 0; j < len(groups); j++ {
			if config.Measurements[i].Name == groups[j] {
                m := Measurement{
                    Metrics: config.Measurements[i].Metrics,
                    Labels: config.Measurements[i].Labels,
                }
                measurements = append(measurements, m)
			}
		}
	}
	return measurements
}

// get_combinations
// itterate measurements and find available instances/nodes reporting those metrics
func get_combinations(measurements []Measurement, prom string) []MetricInstanceCombination {
	var combos []MetricInstanceCombination
    for _, measurement := range measurements {
        for _, metric := range measurement.Metrics {
            response, err := http.Get(prom + "/api/v1/query?query=" + metric)
            if err != nil {
                log.Fatalf("Error querying prometheus: %v", err)
            }
            buf, _ := ioutil.ReadAll(response.Body)
            var resp InstanceResponse
            json.Unmarshal(buf, &resp)
            var combo MetricInstanceCombination
            combo.Metric = metric
            var instanceSet map[string]int = make(map[string]int)
            for j := 0; j < len(resp.Data.Result); j++ {
                res := resp.Data.Result[j].Metric
                if val, ok := res["instance"]; ok {
                    instanceSet[val] = 0
                }
            }
            for k := range instanceSet {
                combo.Instances = append(combo.Instances, k)
            }
            combo.Labels = measurement.Labels
            combos = append(combos, combo)
        }
    }
	return combos
}

func get_data_response(metric string, instance string, start string, end string, step string, prom string) MainResponse {
	query_str := prom + "/api/v1/query_range?query=" + metric + "{instance=\"" + instance + "\"}&start=" + start + "&end=" + end + "&step=" + step
	response, _ := http.Get(query_str)
	buf, _ := ioutil.ReadAll(response.Body)
	var resp MainResponse
    err := json.Unmarshal(buf, &resp)
    if err != nil {
        log.Fatalf("Error parsing response from prom: %v\n", err)
    }
	return resp
}


func conv_time_record(t string) string {
	var sec_nsec []string = strings.Split(t, ".")
	sec, _ := strconv.ParseInt(sec_nsec[0], 10, 64)
	var nsec int64 = 0
	if len(sec_nsec) > 1 {
		n, _ := strconv.ParseInt(sec_nsec[1], 10, 64)
		nsec = n
	}
	return time.Unix(sec, nsec).Format("2006-01-02 15:04:05.000000")
}

// extract values from returned prometheus data
func parse_result(res MainDataResult, metrics string, instance string, labels []string) (*ParsedMetrics) {
    sort.Strings(labels)
    pm := &ParsedMetrics {
        Metric: res.Metric["__name__"],
        Instance: res.Metric["instance"],
        Labels: labels,
        DataPoints: make([]DataPoint, len(res.Values)),
    }
    for i, data := range res.Values {
        dp := DataPoint {
            Timestamp: fmt.Sprintf("%f", data[0].(float64)),
            Value: data[1].(string),
            Labels: make(map[string]string, len(labels)),
        }
        // collect label values
        for _, l := range labels {
            dp.Labels[l] = res.Metric[l]
        }
        pm.DataPoints[i] = dp
    }
    return pm
}

// Metrics with differnt label values are managed as seperate ParsedMetrics,
// this function combines them before they are passed to write_csv
func combine_parsed_metrics(parsed []*ParsedMetrics) *ParsedMetrics {
    // copy first and then copy arrays of each other to first
    p := parsed[0]
    for i:=1;i<len(parsed);i++ {
       p.DataPoints = append(p.DataPoints, parsed[i].DataPoints...)
    }
    return p
}


func parse_response(resp MainResponse, metric string, instance string, labels []string, printout string, dir string, annotation string, start string, end string) {
    var parsed []*ParsedMetrics
    
    for _, res := range resp.Data.Result {
		pm := parse_result(res, metric, instance, labels)
        parsed = append(parsed, pm)
    }
    log.Printf("Parsed: %+v\n", parsed)
    combined := combine_parsed_metrics(parsed)
    write_csv(instance, metric, combined, dir, annotation)
}

func write_csv(instance string, metric string, data *ParsedMetrics, dir string, annotation string) {
	var filename string = dir + "/" + annotation + "/" + metric
	pretty.Println(filename)
	os.MkdirAll(filename, os.ModePerm)
	instance = strings.ReplaceAll(strings.ReplaceAll(instance, ":", "_"), ".", "_")
	csvfile, err := os.Create(filename + "/" + instance + ".csv")
	if err != nil {
		panic(err)
	}
	csvwriter := csv.NewWriter(csvfile)
    for _, row := range data.GetRows() {
        err = csvwriter.Write(row)
        if err != nil {
            panic(err)
        }
    }
	csvwriter.Flush()
	csvfile.Close()
}

func convert_time(t string) string {
	const long_form = "2006-01-02 15:04:05.999"
	tim, _ := time.Parse(long_form, t)
	return strconv.FormatInt(tim.Unix(), 10)
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    prom := flag.String("prom", "http://192.168.1.2:8080", "hostname and port of prometheus instance")
	conf := flag.String("conf", "./prom_dump.conf", "path to the configuration file")
	dir := flag.String("dir", "./dump", "directory to save data dumps")
	annotation := flag.String("annotation", "job_name", "annotation to attach to path")
	metric_group := flag.String("group", "turbostat,perf", "groups of metrics to pull separated by commas")
	metric_start := flag.String("start", "2019-12-08 13:30:25.000", "Start of range")
	metric_end := flag.String("end", "2019-12-08 14:30:25.000", "Start of range")
	metric_step := flag.String("step", "1s", "Query step to, i.e. '1s', '500ms', '1m', '1h' etc.")
	flag.Parse()
	//var config []string = get_config(*metric_group, *conf)
	var config []Measurement = get_config(*metric_group, *conf)
	var combos []MetricInstanceCombination = get_combinations(config, *prom)
    fmt.Printf("Combos: %v\n", combos)

	length := 0
	for i := 0; i < len(combos); i++ {
		length = length + len(combos[i].Instances)
	}
	var wg sync.WaitGroup
	wg.Add(length)

    log.Printf("About to start: combos len: %d\n", length)

	for i := 0; i < len(combos); i++ {
		for j := 0; j < len(combos[i].Instances); j++ {
			//go func(i int, j int) {
			func(i int, j int) {
				parse_response(
					get_data_response(
						combos[i].Metric,
						combos[i].Instances[j],
						convert_time(*metric_start),
						convert_time(*metric_end),
						*metric_step,
						*prom),
					combos[i].Metric,
					combos[i].Instances[j],
                    combos[i].Labels,
					"Metric: "+
						fmt.Sprintf("%d", i+1)+
						"/"+
						fmt.Sprintf("%d", len(combos))+
						" Instance: "+
						fmt.Sprintf("%d", j+1)+
						"/"+
						fmt.Sprintf("%d", len(combos[i].Instances)),
					*dir,
					*annotation,
					*metric_start,
					*metric_end)
				wg.Done()
			}(i, j)
		}
	}

	wg.Wait()
}
