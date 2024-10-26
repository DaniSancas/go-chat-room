# Description: Run all Go tests

# 1st argument: Go project directory, default is current directory
if [ -z "$1" ]; then
    GO_PROJECT_DIR=$(pwd)
else
    GO_PROJECT_DIR=$(realpath $1)
fi
echo "Go project directory: $GO_PROJECT_DIR"
echo "WARNING: the above path is where the go.mod file is expected!"

# Get go.mod file and extract module name
GO_MOD_FILE=$(find $GO_PROJECT_DIR -name go.mod)
GO_MODULE_NAME=$(cat $GO_MOD_FILE | grep module | awk '{print $2}')
echo "Go module name: $GO_MODULE_NAME"

# Default exit code is 0
EXIT_CODE=0

# Get all subdirectories in project dir
# For each subdirectory, execute the tests concatenating the module and relative path
SUBDIRS=$(find $GO_PROJECT_DIR -mindepth 1 -type d -not -path '*/\.*')
for SUBDIR in $SUBDIRS; do
    RELATIVE_PATH=$(realpath --relative-to=$GO_PROJECT_DIR $SUBDIR)
    # If the relative path doesn't have a go file, skip
    if [ -z "$(find $SUBDIR -maxdepth 1 -name '*.go')" ]; then
        continue
    fi
    
    # Jump to project dir and run tests
    (cd $GO_PROJECT_DIR && go test $GO_MODULE_NAME/$RELATIVE_PATH)
    # If test failed, set exit code to 1
    if [ $? -ne 0 ]; then
        EXIT_CODE=1
    fi
done

exit $EXIT_CODE
