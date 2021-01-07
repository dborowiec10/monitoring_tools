#!/bin/bash

# BASE_PATH="/home/damian/Desktop/gpu-monitoring/prom_dump_csv/dump"
# # enable this for single-job import
# job=$BASE_PATH/vgg19_bs96
# for field in "${job}"/*; do
#     for file in "${field}"/*; do
#         ./conv_to_line_proto --path="$file" --tag="${job##*/}" --measurement="${field##*/}" --dump="./dump"
#     done
# done



# BASE_PATH="../prom_dump_csv/dump"
# BASE_PATH="../mysql_dump_csv/data"

# BASE_PATH="/home/damian/work/DL-GPU-Energy-Project-Experiment-Data/stage_9/prometheus/3600"
BASE_PATH="/home/damian/work/DL-GPU-Energy-Project-Experiment-Data/stage_9/tektronix/3600"

# enable this for multi-job import
for job in $BASE_PATH/*; do
    for field in "${job}"/*; do
        for file in "${field}"/*; do
            ./conv_to_line_proto --path="$file" --tag="${job##*/}_3600" --measurement="${field##*/}" --dump="./dump"
        done
    done
done