all:
	go build -o rak831

rpi:
	GOOS=linux GOARCH=arm GOARM=7 go build -o rak831rpi