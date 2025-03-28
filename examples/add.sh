#!/bin/bash

# Get the values from environment variables
if [ -z "$a" ] || [ -z "$b" ]; then
  echo "Error: Missing required parameters 'a' or 'b'"
  exit 1
fi

# Try to convert to integers
a_val=$(($a))
b_val=$(($b))

# Perform the addition
result=$(($a_val + $b_val))

# Return the result
echo "The sum of $a and $b is $result" 