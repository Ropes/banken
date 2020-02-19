TARGET=banken

all: $(TARGET)

$(TARGET): build grant-capture run

build: 
	go build -o $(TARGET) ./cmd/banken/main.go

grant-capture:
	sudo setcap cap_net_raw,cap_net_admin=eip $(TARGET)

run:
	./$(TARGET) monitor

go-test-banken:
	go test -c ./cmd/banken/cmd -o $(TARGET).test
	sudo setcap cap_net_raw,cap_net_admin=eip $(TARGET).test
	./$(TARGET).test

go-test-race:
	go test -race ./pkg/traffic

clean:
	go clean -i 
	rm -q $(TARGET).test
	rm -q $(TARGET)
