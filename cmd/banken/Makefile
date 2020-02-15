TARGET=banken

all: $(TARGET)

$(TARGET): build

build:
	go build -o $(TARGET) main.go

grant-capture:
	sudo setcap cap_net_raw,cap_net_admin=eip $(TARGET)

clean:
	go clean -i 