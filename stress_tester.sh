#!/bin/bash

# Define the number of times to run the Python script
RUN_COUNT=5
TEST_COUNT=100
# Define the folder to save the outputs
OUTPUT_FOLDER="tests/stress_tests"

# Create the output folder if it doesn't exist
mkdir -p $OUTPUT_FOLDER

get_real_time() {
    # Use `time` command and capture the output
    { time $1 ; } 2>&1 | grep real | awk '{print $2}'
}

# Run the Python script the specified number of times
for ((i=1; i<=RUN_COUNT; i++))
do
    total_time=0
    # Define the argument for this run
    ARGUMENT=$((i * 10000))
    mkdir -p $OUTPUT_FOLDER/size_$i
    for ((j=1; j<=TEST_COUNT; j++))
    do
        mkdir -p $OUTPUT_FOLDER/size_$i/test_$j
        TEST_PATH=$OUTPUT_FOLDER/size_$i/test_$j/test_$j.tri

        COMMAND="./trivil_llvm $TEST_PATH"

        REAL_TIME=$(get_real_time "$COMMAND")

        #echo "DEBUG: REAL_TIME=$REAL_TIME"

        minutes=$(echo $REAL_TIME | awk -F'm' '{print $1}')
        seconds=$(echo $REAL_TIME | awk -F'm' '{print $2}' | awk -F's' '{print $1}')
        seconds=$(echo $seconds | sed 's/,/./')

        # Debugging output
        #echo "DEBUG: minutes=$minutes, seconds=$seconds"

        # Convert time to total seconds
        time_in_seconds=$(echo "$minutes * 60 + $seconds" | bc -l 2>&1)
        #echo "DEBUG: time_in_seconds=$time_in_seconds"
        total_time=$(echo "$total_time + $time_in_seconds" | bc -l 2>&1)
    done
    average_time=$(echo "$total_time / $TEST_COUNT" | bc -l)
    echo "size: $i, average time: $average_time, total time: $total_time"
    # Run the Python script and save the output to a file
done

