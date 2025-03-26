#!/bin/bash

check_go() {
  if ! command -v go &> /dev/null; then
    printf "Go is not installed on your system.\n"
    read -p "Would you like to install Go now? (y/n): " choice
    
    case "$choice" in
      y|Y)
        printf "Installing Go...\n"
        
        # Detect OS
        case "$(uname -s)" in
          Darwin)
            if command -v brew &> /dev/null; then
              brew install go
            else
              printf "Homebrew not found. Please install Go manually from https://golang.org/dl/\n"
              exit 1
            fi
            ;;
          Linux)
            if command -v apt-get &> /dev/null; then
              sudo apt-get update && sudo apt-get install -y golang-go
            elif command -v yum &> /dev/null; then
              sudo yum install -y golang
            else
              printf "Package manager not found. Please install Go manually from https://golang.org/dl/\n"
              exit 1
            fi
            ;;
          MINGW*|MSYS*|CYGWIN*)
            printf "On Windows, please install Go manually from https://golang.org/dl/\n"
            exit 1
            ;;
          *)
            printf "Unsupported operating system. Please install Go manually from https://golang.org/dl/\n"
            exit 1
            ;;
        esac
        
        # Verify installation
        if ! command -v go &> /dev/null; then
          printf "Go installation failed.\n"
          exit 1
        fi
        printf "Go has been installed successfully!\n"
        ;;
      *)
        printf "Go installation skipped. Please install Go before building.\n"
        exit 1
        ;;
    esac
  else
    printf "Go is already installed. Version: %s\n" "$(go version)"
    return 0
  fi
}

check_go