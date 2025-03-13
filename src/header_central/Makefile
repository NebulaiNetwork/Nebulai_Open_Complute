SRC=`find ./src -type f -name "*.go"`
TAR=header_central
BUILD_CMD=go build


all:
	${BUILD_CMD} -o ${TAR} ${SRC}

clean:
	rm -rf ${TAR}
