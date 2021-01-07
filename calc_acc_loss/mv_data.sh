#!/bin/bash

for filename in ./*.txt; do
    mkdir ${filename%.txt}
    mv $filename ${filename%.txt}/$filename
done