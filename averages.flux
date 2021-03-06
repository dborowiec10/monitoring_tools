
from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "dcgm_gpu_utilization")
  |> filter(fn: (r) => r.gpu == "0")
  |> mean()
  |> yield(name: "gpu_util_mean")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "dcgm_gpu_utilization")
  |> filter(fn: (r) => r.gpu == "0")
  |> stddev()
  |> yield(name: "gpu_util_stddev")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "tektronix_power_watts")
  |> mean()
  |> yield(name: "machine_power_mean")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "tektronix_power_watts")
  |> stddev()
  |> yield(name: "machine_power_stddev")
  
from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "turbostat_pkgwatt")
  |> filter(fn: (r) => r.summary == "true")
  |> mean()
  |> yield(name: "cpu_power_mean")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "turbostat_pkgwatt")
  |> filter(fn: (r) => r.summary == "true")
  |> stddev()
  |> yield(name: "cpu_power_stddev")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "dcgm_power_usage")
  |> filter(fn: (r) => r.gpu == "0")
  |> mean()
  |> yield(name: "gpu_power_mean")

from(bucket: "gpu_monitoring")
  |> range(start: 2020-01-22T13:42:33Z, stop: 2020-01-22T13:46:02Z)
  |> filter(fn: (r) => r._measurement == "dcgm_power_usage")
  |> filter(fn: (r) => r.gpu == "0")
  |> stddev()
  |> yield(name: "gpu_power_stdev")


