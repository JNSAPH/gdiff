EXECUTABLE = gdiff
BUILD_DIR = bin
RELEASE_DIR = release

.PHONY: build run dev check tui-studio release clean

build:
	go build -o $(BUILD_DIR)/$(EXECUTABLE)

run: build
	./$(BUILD_DIR)/$(EXECUTABLE)

dev: build
	./$(BUILD_DIR)/$(EXECUTABLE) --dev

check:
	go build ./...
	go vet ./...
	gofmt -l .

tui-studio:
	docker run --name tui-studio -p 8080:80 javieralonso716/tui-studio-web:latest

release: build
	rm -rf $(RELEASE_DIR)
	mkdir -p $(RELEASE_DIR)
	cp $(BUILD_DIR)/$(EXECUTABLE) $(RELEASE_DIR)/$(EXECUTABLE)

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(RELEASE_DIR)
