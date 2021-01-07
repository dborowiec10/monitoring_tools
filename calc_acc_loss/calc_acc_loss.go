package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Header struct {
	job_type      string
	job_name      string
	model         string
	dataset       string
	epochs        string
	learning_rate string
	batch_size    string
	device        string
	stage         string
}

func load_log(log_path string) (Header, []float64, []float64, []float64) {

	job_type := ""
	job_name := ""
	model := ""
	dataset := ""
	epochs := ""
	learning_rate := ""
	batch_size := ""
	device := ""

	losses := []float64{}
	accuracies := []float64{}
	speeds := []float64{}

	fmt.Printf("Loading Log: %s\n", log_path)

	file, err := os.Open(log_path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()

		if strings.Contains(line, ": epoch ") {
			splits := strings.Split(line, " ")
			if loss, err := strconv.ParseFloat(strings.Trim(splits[7], ":,\n "), 64); err == nil {
				losses = append(losses, loss)
			}
			if acc, err := strconv.ParseFloat(strings.Trim(splits[9], ":,\n "), 64); err == nil {
				accuracies = append(accuracies, acc)
			}
			if s, err := strconv.ParseFloat(strings.Trim(splits[11], ":,\n "), 64); err == nil {
				speeds = append(speeds, s)
			}
		} else if strings.Contains(line, ": RUN:") {
			splits := strings.Split(line, " ")
			job_name = strings.Trim(splits[3], ":,\n ")
			job_type = strings.Trim(splits[5], ":,\n ")
			model = strings.Trim(splits[7], ":,\n ")
			dataset = strings.Trim(splits[9], ":,\n ")
		} else if strings.Contains(line, ": META:") {
			splits := strings.Split(line, " ")
			epochs = strings.Trim(splits[4], ":,\n ")
			learning_rate = strings.Trim(splits[6], ":,\n ")
			batch_size = strings.Trim(splits[8], ":,\n ")
		} else if strings.Contains(line, ": DEVICE:") {
			splits := strings.Split(line, " ")
			device = strings.Trim(splits[3], ":,\n ")
		}

	}

	h := Header{
		job_type:      job_type,
		job_name:      job_name,
		model:         model,
		dataset:       dataset,
		epochs:        epochs,
		learning_rate: learning_rate,
		batch_size:    batch_size,
		device:        device}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return h, losses, accuracies, speeds
}

func write(csv_path string, stage string, job_folder string, header Header, losses []float64, accuracies []float64, speeds []float64) {
	file, err := os.OpenFile(csv_path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(file)

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	file.Close()

	records[0] = append(records[0], header.job_type)
	records[0] = append(records[0], header.job_type)
	records[0] = append(records[0], header.job_type)

	records[1] = append(records[1], header.job_name)
	records[1] = append(records[1], header.job_name)
	records[1] = append(records[1], header.job_name)

	records[2] = append(records[2], job_folder)
	records[2] = append(records[2], job_folder)
	records[2] = append(records[2], job_folder)

	records[3] = append(records[3], header.model)
	records[3] = append(records[3], header.model)
	records[3] = append(records[3], header.model)

	records[4] = append(records[4], header.dataset)
	records[4] = append(records[4], header.dataset)
	records[4] = append(records[4], header.dataset)

	records[5] = append(records[5], header.epochs)
	records[5] = append(records[5], header.epochs)
	records[5] = append(records[5], header.epochs)

	records[6] = append(records[6], header.learning_rate)
	records[6] = append(records[6], header.learning_rate)
	records[6] = append(records[6], header.learning_rate)

	records[7] = append(records[7], header.batch_size)
	records[7] = append(records[7], header.batch_size)
	records[7] = append(records[7], header.batch_size)

	records[8] = append(records[8], header.device)
	records[8] = append(records[8], header.device)
	records[8] = append(records[8], header.device)

	records[9] = append(records[9], stage)
	records[9] = append(records[9], stage)
	records[9] = append(records[9], stage)

	records[10] = append(records[10], "losses")
	records[10] = append(records[10], "accuracies")
	records[10] = append(records[10], "speeds")

	rec_length := len(records)
	loss_length := len(losses)

	// prepare empty rows
	if loss_length > rec_length-11 {
		diff := (loss_length - (rec_length - 11))
		for i := 0; i < diff; i++ {
			records = append(records, make([]string, 0))
		}
	}

	for j := 0; j < len(records); j++ {
		if j > 10 && len(records[j]) < 1 {
			records[j] = append(records[j], "")
		}
		if j > 10 && len(records[j]) < len(records[0]) {
			records[j] = append(records[j], "")
			records[j] = append(records[j], "")
			records[j] = append(records[j], "")
		}
	}

	for j := 11; j < len(records); j++ {
		if j-11 < len(losses) {
			length := len(records[j])
			records[j][length-3] = fmt.Sprintf("%.4f", losses[j-11])
			records[j][length-2] = fmt.Sprintf("%.4f", accuracies[j-11])
			records[j][length-1] = fmt.Sprintf("%.4f", speeds[j-11])
		}
	}

	write_file, err2 := os.OpenFile(csv_path, os.O_WRONLY|os.O_CREATE, 0777)
	if err2 != nil {
		log.Fatal(err2)
	}

	writer := csv.NewWriter(write_file)
	err_wrt := writer.WriteAll(records)
	if err_wrt != nil {
		log.Fatal(err_wrt)
	}
	writer.Flush()
	write_file.Close()
}

func main() {
	_stage := flag.String("stage", "1", "stage number")
	_in := flag.String("in", "/data/file.log", "path to log file")
	_out := flag.String("out", "/data/file.csv", "path where to data in csv format")
	_job_folder := flag.String("job_folder", "job...", "job...")

	flag.Parse()

	h, losses, accuracies, speeds := load_log(*_in)

	write(*_out, *_stage, *_job_folder, h, losses, accuracies, speeds)
}
