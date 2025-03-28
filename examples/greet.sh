#!/bin/bash

# Get the values from environment variables
if [ -z "$name" ]; then
  echo "Error: Missing required parameter 'name'"
  exit 1
fi

# Set default values if not provided
if [ -z "$greeting" ]; then
  greeting="Hello"
fi

if [ -z "$formal" ]; then
  formal=false
fi

# Customize greeting based on formal flag
if [ "$formal" = "true" ]; then
  title="Mr./Ms."
  message="${greeting}, ${title} ${name}. How may I assist you today?"
else
  message="${greeting}, ${name}! Nice to meet you!"
fi

# Return the greeting
echo "$message" 