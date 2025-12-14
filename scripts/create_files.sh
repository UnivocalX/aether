#!/bin/bash
# create_files.sh

# Check if arguments are provided
if [ $# -lt 2 ]; then
    echo "Usage: $0 <output_directory> <number_of_files>"
    echo "Example: $0 ./myfiles 10000"
    exit 1
fi

OUTPUT_DIR="$1"
NUM_FILES="$2"

# Validate number is positive integer
if ! [[ "$NUM_FILES" =~ ^[0-9]+$ ]]; then
    echo "Error: Number of files must be a positive integer"
    exit 1
fi

if [ "$NUM_FILES" -lt 1 ]; then
    echo "Error: Number of files must be at least 1"
    exit 1
fi

echo "Creating $NUM_FILES empty files in '$OUTPUT_DIR'..."

# Create directory (with parents if needed)
mkdir -p "$OUTPUT_DIR"

# Check if directory was created successfully
if [ ! -d "$OUTPUT_DIR" ]; then
    echo "Error: Failed to create directory '$OUTPUT_DIR'"
    exit 1
fi

# Count existing files
EXISTING_FILES=$(find "$OUTPUT_DIR" -maxdepth 1 -type f | wc -l)
echo "Found $EXISTING_FILES existing files in directory"

# Create files
for ((i=1; i<=NUM_FILES; i++)); do
    touch "$OUTPUT_DIR/file_$i.txt"
    
    # Show progress every 10%
    if [ $((i % (NUM_FILES / 10))) -eq 0 ] && [ "$NUM_FILES" -ge 10 ]; then
        PERCENT=$((i * 100 / NUM_FILES))
        echo "Progress: $PERCENT% ($i/$NUM_FILES)"
    fi
done

echo "Done! Created $NUM_FILES files in '$OUTPUT_DIR'"

# Final count
TOTAL_FILES=$(find "$OUTPUT_DIR" -maxdepth 1 -type f | wc -l)
echo "Total files in directory: $TOTAL_FILES"