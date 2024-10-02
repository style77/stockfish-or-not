#!/bin/bash

get_architecture() {
    arch=$(uname -m)
    os=$(uname -s)

    if [[ "$os" == "Linux" ]]; then
        if [[ "$arch" == "x86_64" ]]; then
            if grep -q avx2 /proc/cpuinfo; then
                echo "linux_avx2"
            elif grep -q sse4_1 /proc/cpuinfo && grep -q popcnt /proc/cpuinfo; then
                echo "linux_popcnt"
            else
                echo "unsupported_linux"
            fi
        elif [[ "$arch" == "armv8" || "$arch" == "aarch64" ]]; then
            echo "armv8"
        elif [[ "$arch" == "armv7l" ]]; then
            echo "armv7"
        else
            echo "unsupported_linux"
        fi
    elif [[ "$os" == "Windows_NT" ]]; then
        if [[ "$arch" == "x86_64" ]]; then
            if grep -q avx2 /proc/cpuinfo; then
                echo "windows_avx2"
            elif grep -q sse4_1 /proc/cpuinfo && grep -q popcnt /proc/cpuinfo; then
                echo "windows_popcnt"
            else
                echo "unsupported_windows"
            fi
        else
            echo "unsupported_windows"
        fi
    else
        echo "unsupported_os"
    fi
}

download_stockfish() {
    arch=$(get_architecture)

    case "$arch" in
        linux_avx2)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-ubuntu-x86-64-avx2.tar"
            ;;
        linux_popcnt)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-ubuntu-x86-64-sse41-popcnt.tar"
            ;;
        windows_avx2)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-windows-x86-64-avx2.zip"
            ;;
        windows_popcnt)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-windows-x86-64-sse41-popcnt.zip"
            ;;
        armv8)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-android-armv8.tar"
            ;;
        armv7)
            stockfish_url="https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-android-armv7.tar"
            ;;
        *)
            echo "Unsupported architecture or operating system: $arch"
            exit 1
            ;;
    esac

    echo "Downloading Stockfish from: $stockfish_url"

    wget -O stockfish_package "$stockfish_url" || {
        echo "Failed to download Stockfish. Check your internet connection."
        exit 1
    }

    if [[ "$stockfish_url" == *.tar ]]; then
        tar -xf stockfish_package -C stockfish_folder || {
            echo "Failed to extract Stockfish."
            exit 1
        }
    elif [[ "$stockfish_url" == *.zip ]]; then
        unzip stockfish_package -d stockfish_folder || {
            echo "Failed to unzip Stockfish."
            exit 1
        }
    fi

    stockfish_binary=$(find stockfish_folder -type f -name 'stockfish*')
    if [[ -z "$stockfish_binary" ]]; then
        echo "Could not find the Stockfish binary in the downloaded package."
        exit 1
    fi

    mv "$stockfish_binary" stockfish || {
        echo "Failed to rename Stockfish binary."
        exit 1
    }

    rm -rf stockfish_folder stockfish_package
}

move_to_api() {
    api_dir="./api"

    if [[ ! -d "$api_dir" ]]; then
        echo "API directory $api_dir does not exist."
        exit 1
    fi

    mv stockfish "$api_dir" || {
        echo "Failed to move Stockfish to $api_dir."
        exit 1
    }

    echo "Stockfish binary successfully moved to $api_dir"
}

mkdir -p stockfish_folder

download_stockfish
move_to_api