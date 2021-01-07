#!/bin/bash

START_DATE=$1
START_TIME=$2
END_DATE=$3
END_TIME=$4

echo "
from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"dcgm_gpu_utilization\")
  |> filter(fn: (r) => r.gpu == \"0\")
  |> mean()
  |> yield(name: \"gpu_util_mean\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"dcgm_gpu_utilization\")
  |> filter(fn: (r) => r.gpu == \"0\")
  |> stddev()
  |> yield(name: \"gpu_util_stddev\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"tektronix_power_watts\")
  |> mean()
  |> yield(name: \"machine_power_mean\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"tektronix_power_watts\")
  |> stddev()
  |> yield(name: \"machine_power_stddev\")
  
from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"turbostat_pkgwatt\")
  |> filter(fn: (r) => r.summary == \"true\")
  |> mean()
  |> yield(name: \"cpu_power_mean\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"turbostat_pkgwatt\")
  |> filter(fn: (r) => r.summary == \"true\")
  |> stddev()
  |> yield(name: \"cpu_power_stddev\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"dcgm_power_usage\")
  |> filter(fn: (r) => r.gpu == \"0\")
  |> mean()
  |> yield(name: \"gpu_power_mean\")

from(bucket: \"gpu_monitoring\")
  |> range(start: ${START_DATE}T${START_TIME}Z, stop: ${END_DATE}T${END_TIME}Z)
  |> filter(fn: (r) => r._measurement == \"dcgm_power_usage\")
  |> filter(fn: (r) => r.gpu == \"0\")
  |> stddev()
  |> yield(name: \"gpu_power_stdev\")

" > averages.flux


