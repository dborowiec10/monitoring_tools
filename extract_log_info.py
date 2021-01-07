from absl import app
from absl import flags
from datetime import datetime
from datetime import timedelta
import os
import csv



FLAGS = flags.FLAGS
flags.DEFINE_string('log', '../DL-GPU-Energy-Project-Experiment-Data/stage_9/logs/3600/mnasnet0_5_resnet18_vgg19_bs96', 'The path to the log file dir')

# def roundTime(dt=None, roundTo=60):
#    if dt == None : dt = datetime.now()
#    seconds = (dt.replace(tzinfo=None) - dt.min).seconds
#    rounding = (seconds+roundTo/2) // roundTo * roundTo
#    return dt + timedelta(0,rounding-seconds,-dt.microsecond)


# main program function
def main(argv):
    del argv

    gpu_durations = []
    log_periods = []
    files = []

    earliest_time = None
    latest_time = None

    for r, d, f in os.walk(FLAGS.log):
        fs = []
        print(r)
        for ff in f:
            fs.append(ff)
        fs = sorted(fs)
        print(fs)
        files.append((r, fs))

    for r, f in files:
        for file in f:
            with open(r + "/" + file) as log_file:
                log_start = None
                log_end = None
                for line in log_file:
                    if ": RUN:" in line:
                        splits = line.strip().split()
                        log_start = datetime.strptime(splits[0] + " " + splits[1].strip(':'), '%Y-%m-%d %H:%M:%S')
                    elif ": job time: " in line:
                        splits = line.strip().split()
                        log_end = datetime.strptime(splits[0] + " " + splits[1].strip(':'), '%Y-%m-%d %H:%M:%S')
                        gpu_durations.append(splits[4])
                
                if earliest_time == None:
                    earliest_time = log_start
                else:
                    if log_start < earliest_time:
                        earliest_time = log_start
                
                if latest_time == None:
                    latest_time = log_end
                else:
                    if log_end > latest_time:
                        latest_time = log_end

                log_periods.append(log_start.strftime("%Y-%m-%d %H:%M:%S"))
                log_periods.append(log_end.strftime("%Y-%m-%d %H:%M:%S"))
                log_periods.append(str(log_end - log_start))

    log_periods.extend(gpu_durations)

    with open('out.csv', mode='w') as fil:
        fil_writer = csv.writer(fil, delimiter=',', quotechar='"', quoting=csv.QUOTE_MINIMAL)
        fil_writer.writerow(log_periods)

    print("Collection Start: " + str(earliest_time - timedelta(seconds=30)))
    print("Collection End: " + str(latest_time + timedelta(seconds=30)))


# entrypoint
if __name__ == "__main__":
    app.run(main)