#!/usr/bin/env python3
"""
project2txt.py

Python script to scan a project's directory structure for common programming files and export it to a text file.
Respects .gitignore rules if present, handles subdirectories, and preserves file contents.
Only includes paths that contain common programming and project files.

Author: ph33nx
GitHub: https://gist.github.com/ph33nx/12ce315ef6dbcf9ef1d01f5371af4a3d

Features:
- Scans for common programming files in a directory and its subdirectories.
- Supported file types include a wide range of file extensions for languages such as Python, JavaScript, Java, C/C++,
  Ruby, PHP, HTML, CSS, configuration files, and more.
- Also supports special filenames (without an extension) such as Dockerfile, Makefile, CMakeLists.txt, etc.
- Excludes files/directories specified in .gitignore.
- Always ignores `.git` and hidden files/folders (e.g., `.env`, `.vscode`).
- Skips binary files like images and executables.
- Outputs a structured text file (YAML formatted) with project structure, file contents, and a placeholder tasks section.
- Dynamically installs PyYAML if it's not already installed.
- Enforces path input; shows help if no path is provided.

Usage:
    python3 project2txt.py [project_path]

Example:
    python3 project2txt.py .
"""

import os
import fnmatch
import argparse
import logging
import subprocess
import sys

# Set up logging
logging.basicConfig(
    format="%(asctime)s - %(levelname)s - %(message)s", level=logging.INFO
)

# Attempt to import PyYAML
try:
    import yaml
except ModuleNotFoundError:
    logging.warning("PyYAML module not found. Attempting to install it...")
    try:
        subprocess.check_call([sys.executable, "-m", "pip", "install", "pyyaml"])
        import yaml

        logging.info("PyYAML successfully installed!")
    except Exception as e:
        logging.critical(f"Failed to install PyYAML: {e}")
        sys.exit(1)

# Constant: List of files and folders to ignore by default.
IGNORE = [
    ".git",  # Git folder
    ".*",  # Hidden files/folders (e.g., .env, .vscode)
    "node_modules",  # Node dependencies
    "bower_components",
    "vendor",  # PHP, Ruby, or other vendor libraries
    "dist",
    "build",
    "venv",
    "env",
    "__pycache__",
    "target",
    "out",
    "logs",
    "tmp",
]

# Define allowed file extensions for common programming and project files.
ALLOWED_EXTENSIONS = {
    # Scripting and dynamic languages
    ".py",
    ".pyw",
    ".rb",
    ".pl",
    ".pm",
    ".php",
    ".phtml",
    ".js",
    ".mjs",
    ".jsx",
    ".ts",
    ".tsx",
    # Markup languages and documentation
    ".html",
    ".htm",
    ".xml",
    ".xhtml",
    ".md",
    ".markdown",
    # Stylesheets
    ".css",
    ".scss",
    ".sass",
    ".less",
    # Compiled languages and header files
    ".c",
    ".cpp",
    ".cxx",
    ".cc",
    ".h",
    ".hpp",
    ".hxx",
    ".hh",
    ".cs",
    ".java",
    ".kt",
    ".kts",
    ".rs",
    # Configuration and data files
    ".json",
    ".yml",
    ".yaml",
    ".ini",
    ".toml",
    ".properties",
    # Shell and batch scripts
    ".sh",
    ".bash",
    ".zsh",
    ".ksh",
    ".bat",
    ".cmd",
    # R language
    ".r",
    # Swift, Dart, Go, Lua, Haskell
    ".swift",
    ".dart",
    ".go",
    ".lua",
    ".hs",
    # Scala
    ".scala",
    ".sbt",
    # Objective-C / Objective-C++
    ".m",
    ".mm",
    # Gradle and Groovy
    ".gradle",
    ".groovy",
    # Elixir
    ".ex",
    ".exs",
}

# Define allowed filenames for common project files that do not have extensions
# (all lower-case for case-insensitive matching).
ALLOWED_FILENAMES = {
    "dockerfile",  # Docker configuration
    "makefile",  # Make build scripts
    "cmakelists.txt",  # CMake configuration
    "rakefile",  # Ruby Rake build file
    "gemfile",  # Ruby Gem dependencies
    "vagrantfile",  # Vagrant configuration
    "procfile",  # Heroku/process management
    "go.mod",  # Go module definition
    "go.sum",  # Go module checksums
    "gradlew",  # Gradle wrapper script (Unix)
    "gradlew.bat",  # Gradle wrapper script (Windows)
    "readme",  # Project readme (if not using an extension)
    "license",  # License file
    "changelog",  # Changelog file
}


def parse_gitignore(base_path):
    """
    Parse the .gitignore file in the project root and extract ignore patterns.
    If .gitignore doesn't exist, return an empty list.
    Then, append the default IGNORE constant to ensure common files/folders are excluded.
    """
    gitignore_path = os.path.join(base_path, ".gitignore")
    ignore_patterns = []
    if os.path.exists(gitignore_path):
        logging.info(f"Found .gitignore at {gitignore_path}. Parsing...")
        try:
            with open(gitignore_path, "r") as f:
                ignore_patterns = [
                    line.strip()
                    for line in f
                    if line.strip() and not line.startswith("#")
                ]
        except Exception as e:
            logging.error(f"Error reading .gitignore: {e}")
    else:
        logging.info(".gitignore not found. Proceeding without ignoring files.")

    # Always add default ignore patterns.
    ignore_patterns.extend(IGNORE)
    return ignore_patterns


def is_ignored(path, base_path, ignore_patterns):
    """
    Check if a file or directory should be ignored based on the provided patterns.

    This function also checks if any segment (directory or file name) in the relative path
    exactly matches one of the ignore patterns (when the pattern contains no glob wildcards).
    """
    relative_path = os.path.relpath(path, base_path).replace("\\", "/")
    segments = relative_path.split("/")

    for pattern in ignore_patterns:
        # If pattern is a plain string (no wildcards), check each segment.
        if not any(char in pattern for char in "*?[]"):
            if pattern in segments:
                logging.debug(f"Ignored: {relative_path} (segment match: {pattern})")
                return True

        # Otherwise, perform fnmatch-based pattern matching.
        if fnmatch.fnmatch(relative_path, pattern) or fnmatch.fnmatch(
            relative_path, f"{pattern}/"
        ):
            logging.debug(f"Ignored: {relative_path} (matches pattern: {pattern})")
            return True
    return False


def is_binary(file_path):
    """
    Check if a file is binary by reading its first 1024 bytes.
    """
    try:
        with open(file_path, "rb") as f:
            chunk = f.read(1024)
        return b"\0" in chunk  # Presence of a null byte often indicates binary content.
    except Exception as e:
        logging.error(f"Error checking if file is binary: {file_path}. {e}")
        return True  # Treat unreadable files as binary.


def is_programming_file(filename):
    """
    Determine if a file is a common programming or project file based on its extension or its full name.
    """
    # Check for allowed filenames (case-insensitive).
    if filename.lower() in ALLOWED_FILENAMES:
        return True

    # Extract extension and check against allowed extensions (case-insensitive).
    _, ext = os.path.splitext(filename)
    if ext.lower() in ALLOWED_EXTENSIONS:
        return True

    return False


def read_project_structure(base_path, ignore_patterns):
    """
    Traverse the project directory and read file names and contents for common programming/project files,
    while ignoring files listed in .gitignore, default ignore patterns, or those that are binary.
    Only files that match the allowed criteria are included.
    """
    project_data = {}
    for root, dirs, files in os.walk(base_path):
        # Filter out ignored directories.
        dirs[:] = [
            d
            for d in dirs
            if not is_ignored(os.path.join(root, d), base_path, ignore_patterns)
        ]

        for file in files:
            full_file_path = os.path.join(root, file)

            # Skip files that match ignore patterns.
            if is_ignored(full_file_path, base_path, ignore_patterns):
                continue

            # Only process files that are recognized as common programming/project files.
            if not is_programming_file(file):
                logging.debug(f"Skipping non-programming file: {full_file_path}")
                continue

            # Skip binary files.
            if is_binary(full_file_path):
                logging.info(f"Skipping binary file: {full_file_path}")
                continue

            # Read file contents (UTF-8, ignoring errors).
            relative_path = os.path.relpath(full_file_path, base_path)
            try:
                with open(full_file_path, "r", encoding="utf-8", errors="ignore") as f:
                    content = f.read()
                logging.info(f"Read file: {relative_path}")
            except Exception as e:
                logging.error(f"Error reading file {relative_path}: {e}")
                content = None  # If unreadable, continue with None

            project_data[relative_path] = content
    return project_data


def generate_txt(project_data, output_path):
    """
    Write the project structure, file contents, and placeholder tasks to a text file (YAML formatted) in the root directory.
    The output file is saved with a .txt extension, (you can use .yaml instead, but many LLM's have better support for .txt).
    """
    txt_file_path = os.path.join(output_path, "project_structure.txt")

    # Custom representer to use block style for multi-line strings in YAML.
    def str_presenter(dumper, data):
        if "\n" in data:
            return dumper.represent_scalar("tag:yaml.org,2002:str", data, style="|")
        return dumper.represent_scalar("tag:yaml.org,2002:str", data)

    # Add the custom string representer to PyYAML.
    yaml.add_representer(str, str_presenter)

    try:
        # Build the data dictionary.
        data = {
            "tasks": [{"message": "Example task 1"}],
            "project_structure": list(project_data.keys()),
            "files": [
                {"name": file, "content": content}
                for file, content in project_data.items()
            ],
        }
        # Dump the YAML-formatted data into a .txt file using UTF-8 encoding.
        with open(txt_file_path, "w", encoding="utf-8") as f:
            yaml.dump(
                data, f, default_flow_style=False, sort_keys=False, allow_unicode=True
            )
        logging.info(f"Text file written to {txt_file_path}")
    except Exception as e:
        logging.error(f"Failed to write text file: {e}")


def main():
    # Set up command-line argument parser.
    parser = argparse.ArgumentParser(
        description="Export a project's common programming and project files to a text file."
    )
    parser.add_argument(
        "project_path",
        help="Path to the project directory. Use '.' for the current directory.",
    )
    args = parser.parse_args()

    # Resolve and validate the project path.
    project_path = os.path.abspath(args.project_path)
    if not os.path.isdir(project_path):
        logging.error(f"Invalid project path: {project_path}")
        parser.print_help()
        sys.exit(1)

    logging.info(f"Scanning project directory: {project_path}")

    # Add the script's own file name to the ignore patterns so it won't be processed.
    script_name = os.path.basename(__file__)
    ignore_patterns = parse_gitignore(project_path)
    ignore_patterns.append(script_name)

    # Read the project structure and file contents (only for allowed files).
    project_data = read_project_structure(project_path, ignore_patterns)

    # Generate the text file containing the project structure and file contents.
    generate_txt(project_data, project_path)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        logging.warning("Process interrupted by user.")
        sys.exit(1)
    except Exception as e:
        logging.critical(f"Unexpected error: {e}")
        sys.exit(1)
