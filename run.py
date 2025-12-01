#!/usr/bin/env python

from pathlib import Path
import zipfile
import sys
import os


files = " ".join(map(str, Path().glob("*.go")))
windows_folder = "bin/windows"
linux_folder = "bin/linux"
commands = {
    "run": f"go run {files}"
}


build_commands = {
    "amd64": [
        f"windows;go build -trimpath -ldflags \"-s -w\"  -o {windows_folder}/cpmass.exe {files}",
        f"linux;go build -trimpath -ldflags \"-s -w\"  -o {linux_folder}/cpmass {files}",
    ],
    "386": [
        f"windows;go build -trimpath -ldflags \"-s -w\"  -o {windows_folder}/cpmass_32.exe {files}",
        f"linux;go build -trimpath -ldflags \"-s -w\"  -o {linux_folder}/cpmass_32 {files}",
    ],
}


def create_zip(zip_filename, folder_to_zip):
    zip_filepath = Path(zip_filename)
    folder_path = Path(folder_to_zip)

    with zipfile.ZipFile(zip_filepath, 'w') as zip_file:
        for file in folder_path.rglob('*'):
            if file.is_file() and file.name != zip_filepath.name:
                zip_file.write(file, file.relative_to(folder_path))


if len(sys.argv) >= 2:
    arg = sys.argv[1]
    if arg != "build":
        exit(1)
    
    for arch in build_commands.keys():
        os.environ["GOARCH"] = arch
        for command in build_commands[arch]:
            os_name, command = command.split(";", 1)
            os.environ["GOOS"] = os_name
            os.system(command)
    create_zip(f"{windows_folder}/cpmass_windows.zip", windows_folder)
    create_zip(f"{linux_folder}/cpmass_linux.zip", linux_folder)
else:
    os.system(commands["run"])
        