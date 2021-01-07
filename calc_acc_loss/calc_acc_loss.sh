#!/bin/bash

BASE_PATH="./logs"
# enable this for multi-job import
for stage in $BASE_PATH/*; do
    for job in "${stage}"/*; do
        for file in "${job}"/*; do
            ./calc_acc_loss --in="$file" --stage="${stage##*/}" --out="out.csv" --job_folder="${job##*/}"
        done
    done
done