VENDOR       = vendor
TEMPLATES    = templates
ASSETS       = $(TEMPLATES)/*.html smtp.json
BINDATA      = assets.go
SRCS         = *.go
TARGET       = mailr
DEBUGTARGET  = mailr-debug
SMTP         = smtp.json
BINDATAFLAGS = -pkg main -o $(BINDATA)

default: build

run: TARGET = $(DEBUGTARGET)
run: debug
	./$(TARGET)

debug: BINDATAFLAGS += -debug
debug: TARGET = $(DEBUGTARGET)
debug: build

build: $(TARGET)

clean:
	rm -f $(BINDATA) $(TARGET) $(DEBUGTARGET)

.PHONY: build clean debug run

$(BINDATA): $(TEMPLATES) $(ASSETS)
	go-bindata $(BINDATAFLAGS) $(ASSETS)

$(TARGET): $(VENDOR) $(SRCS) $(BINDATA)
	go build -o $(TARGET) $(SRCS) 

$(VENDOR):
	glide up
