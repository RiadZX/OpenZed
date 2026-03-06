#!/bin/bash

# Enable debug mode to print all executed commands
#set -x

# Function to print messages with timestamps and colors
print_message() {
    local color=$1
    local message=$2
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    echo -e "\e[${color}m[${timestamp}] ${message}\e[0m"
}

# Check if the 7z command is available
if ! command -v 7z &> /dev/null; then
    print_message "31" "7z command not found. Please install 7zip and try again."
    exit 1
fi

# Define the different sets of flags
declare -A flag_sets
flag_sets["Zed"]="-X 'Zed/ui.WhiteLabel=Zed'"

# Output directory for the built binaries
output_dir="builds"
mkdir -p $output_dir

# Define paths to the configuration files
config_dir="./emptyConfigs"
toml_file="$config_dir/config.toml"
csv_file="$config_dir/tasks.csv"
guide_file="$config_dir/guide.md"

# Define target platforms
platforms=("windows/amd64" "darwin/arm64" "darwin/amd64" "linux/amd64")

# Create a temporary directory
tmp_dir=$(mktemp -d)

# Loop through each set of flags and build the application for each platform
for version in "${!flag_sets[@]}"; do
    flags=${flag_sets[$version]}
    for platform in "${platforms[@]}"; do
        os=$(echo $platform | cut -d'/' -f1)
        arch=$(echo $platform | cut -d'/' -f2)
        ext=""
        if [ "$os" = "windows" ]; then
            ext=".exe"
        fi
        output_file="${version}_${os}_${arch}${ext}"
        print_message "34" "Building $version for $os/$arch with flags: $flags"
        if ! GOOS=$os GOARCH=$arch go build -ldflags "$flags -s -w" -o $output_file; then
            print_message "31" "Build failed for $version on $os/$arch"
            exit 1
        fi
        print_message "32" "Build succeeded for $version on $os/$arch: $output_file"


        cp "$toml_file" "$tmp_dir/config.toml"


        # Create a .zip file containing the binary and the configuration files
        zip_file="$output_dir/${version}_${os}_${arch}.zip"
        print_message "34" "Creating .zip file: $zip_file"
        if ! 7z a "$zip_file" "$output_file" "$tmp_dir/config.toml" "$csv_file" "$guide_file" -r; then
            print_message "31" "Failed to create .zip file: $zip_file"
            exit 1
        fi
        print_message "32" ".zip file created: $zip_file"

        # Delete the binary file
        print_message "34" "Deleting binary file: $output_file"
        if ! rm "$output_file"; then
            print_message "31" "Failed to delete binary file: $output_file"
            exit 1
        fi
        print_message "32" "Binary file deleted: $output_file"

        # Remove the temporary config.toml file
        rm "$tmp_dir/config.toml"
    done
done

rm -r "$tmp_dir"

print_message "32" "All builds completed successfully."

# Disable debug mode
set +x